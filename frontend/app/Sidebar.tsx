import { useState, KeyboardEvent, ChangeEvent, useRef } from "react";
import "./Sidebar.css";
import type { SidebarProps } from "@/types";
import * as apiService from "@/services/apiService";
import { toast } from "sonner";

interface SidebarExtendedProps extends SidebarProps {
  roomCode?: string;
  onUploadPdf?: () => void;
  pdfs?: any[];
  activePdfId?: number | null;
  onSelectPdf?: (id: number) => void;
}

function Sidebar({
  documents,
  activeDocId,
  onSelectDocument,
  onCreateDocument,
  isCollapsed,
  onToggle,
  roomCode,
  onUploadPdf,
  pdfs,
  activePdfId,
  onSelectPdf,
}: SidebarExtendedProps) {
  const [newDocTitle, setNewDocTitle] = useState<string>("");
  const [showCreateInput, setShowCreateInput] = useState<boolean>(false);
  const [uploading, setUploading] = useState<boolean>(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleCreate = () => {
    if (newDocTitle.trim()) {
      onCreateDocument(newDocTitle.trim());
      setNewDocTitle("");
      setShowCreateInput(false);
    }
  };

  const handleKeyPress = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleCreate();
    } else if (e.key === "Escape") {
      setShowCreateInput(false);
      setNewDocTitle("");
    }
  };

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setNewDocTitle(e.target.value);
  };

  const handlePdfUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.name.toLowerCase().endsWith(".pdf")) {
      toast.error("Only PDF files are allowed");
      return;
    }

    if (!roomCode) {
      toast.error("Room code is required");
      return;
    }

    setUploading(true);
    try {
      await apiService.uploadPdf(roomCode as string, file);
      toast.success("PDF uploaded successfully!");
      if (onUploadPdf) {
        onUploadPdf();
      }
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    } catch (error: any) {
      const status = (error as any).status;
      const message = error.message || "";

      if (
        status === 403 &&
        message.toLowerCase().includes("github account not linked")
      ) {
        toast.custom(
          () => (
            <div className="flex flex-col gap-2 bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
              <p className="font-medium">
                Link your GitHub account to upload PDFs
              </p>
              <button
                onClick={() => {
                  window.location.href = `${window.location.origin}/auth`;
                }}
                className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded font-medium text-sm w-full"
              >
                Link GitHub Account
              </button>
            </div>
          ),
          { duration: 10000 },
        );
      } else {
        toast.error(error.message || "Failed to upload PDF");
      }
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className={`sidebar ${isCollapsed ? "collapsed" : ""}`}>
      <div className="sidebar-header">
        {!isCollapsed && (
          <>
            <h3>Documents</h3>
            <button
              className="toggle-btn"
              onClick={onToggle}
              title="Collapse sidebar"
            >
              ‹
            </button>
          </>
        )}
        {isCollapsed && (
          <button
            className="toggle-btn collapsed"
            onClick={onToggle}
            title="Expand sidebar"
          >
            ›
          </button>
        )}
      </div>

      {!isCollapsed && (
        <div className="sidebar-content">
          <div className="documents-list">
            {documents.map((doc) => (
              <div
                key={doc.id}
                className={`document-item ${doc.id === activeDocId ? "active" : ""}`}
                onClick={() => onSelectDocument(doc.id)}
              >
                <span className="doc-icon">📄</span>
                <span className="doc-title">{doc.title}</span>
              </div>
            ))}

            {documents.length === 0 && (
              <div className="empty-state">
                <p>No documents yet</p>
                <small>Create your first document</small>
              </div>
            )}
          </div>

          <div className="pdfs-list">
            {pdfs && pdfs.length > 0 && (
              <>
                <div className="pdfs-header">📎 PDFs</div>
                {pdfs.map((pdf) => (
                  <div
                    key={pdf.id}
                    className={`pdf-item ${pdf.id === activePdfId ? "active" : ""}`}
                  >
                    <div
                      className="pdf-item-content"
                      onClick={() => onSelectPdf?.(pdf.id)}
                    >
                      <span className="pdf-icon">📋</span>
                      <span className="pdf-title">{pdf.filename}</span>
                    </div>
                    <button
                      className="pdf-delete-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        apiService
                          .deletePdf(pdf.id)
                          .then(() => {
                            toast.success("PDF deleted successfully!");
                            onUploadPdf?.();
                          })
                          .catch((error: any) => {
                            const status = (error as any).status;
                            const message = error.message || "";

                            if (
                              status === 403 &&
                              message
                                .toLowerCase()
                                .includes("github account not linked")
                            ) {
                              toast.custom(
                                () => (
                                  <div className="flex flex-col gap-2 bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
                                    <p className="font-medium">
                                      Link your GitHub account to delete PDFs
                                    </p>
                                    <button
                                      onClick={() => {
                                        window.location.href = `${window.location.origin}/auth`;
                                      }}
                                      className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded font-medium text-sm w-full"
                                    >
                                      Link GitHub Account
                                    </button>
                                  </div>
                                ),
                                { duration: 10000 },
                              );
                            } else {
                              toast.error(
                                error.message || "Failed to delete PDF",
                              );
                            }
                          });
                      }}
                      title="Delete PDF"
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </>
            )}
          </div>

          <div className="sidebar-footer">
            {!showCreateInput ? (
              <button
                className="create-doc-btn"
                onClick={() => setShowCreateInput(true)}
              >
                + New Document
              </button>
            ) : (
              <div className="create-input-group">
                <input
                  type="text"
                  placeholder="Document title..."
                  value={newDocTitle}
                  onChange={handleChange}
                  onKeyDown={handleKeyPress}
                  autoFocus
                  className="create-input"
                />
                <div className="create-actions">
                  <button className="create-btn-small" onClick={handleCreate}>
                    Create
                  </button>
                  <button
                    className="cancel-btn-small"
                    onClick={() => {
                      setShowCreateInput(false);
                      setNewDocTitle("");
                    }}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf"
              onChange={handlePdfUpload}
              disabled={uploading}
              style={{ display: "none" }}
              id="pdf-file-input"
            />
            <label htmlFor="pdf-file-input" className="upload-pdf-btn">
              {uploading ? "Uploading PDF..." : "Upload PDF"}
            </label>
          </div>
        </div>
      )}
    </div>
  );
}

export default Sidebar;
