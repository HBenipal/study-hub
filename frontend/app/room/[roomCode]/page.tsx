"use client";

import React from "react";
import { useState, useEffect, useRef } from "react";
import { marked } from "marked";
const renderMathInElement: any = require("katex/dist/contrib/auto-render");
import "katex/dist/katex.min.css";
import "../../App.css";
import Sidebar from "@/app/Sidebar";
import { useParams } from "next/navigation";
import * as apiService from "@/services/apiService";
import { applyOperation, transformCursor, computeOperations } from "@/utils/ot";
import type { WSEventData } from "@/types";
import { EditorHeader } from "@/components/EditorHeader";
import { MarkdownEditor } from "@/components/Markdown/MarkdownEditor";
import { MarkdownPreview } from "@/components/Markdown/MarkdownPreview";
import { toast } from "sonner";
import { EllipsisIcon, Pause, Sparkles } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { TooltipArrow } from "@radix-ui/react-tooltip";

marked.setOptions({
  breaks: true,
  gfm: true,
});

// CITATION
// this code is AI assisted. prompt: how to configure marked library so that it escapes html
marked.use({
  hooks: {
    preprocess(markdown) {
      // escape HTML
      return markdown.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    },
  },
});
// END CITATION

export default function Home() {
  const [documents, setDocuments] = useState<any[]>([]);
  const [pdfs, setPdfs] = useState<any[]>([]);
  const [room, setRoom] = useState<string>("");
  const [activeDocId, setActiveDocId] = useState<number | null>(null);
  const [activePdfId, setActivePdfId] = useState<number | null>(null);
  const [content, setContent] = useState("");
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true);

  const [status, setStatus] = useState("disconnected");
  const [statusText, setStatusText] = useState("Connecting...");
  const [userCount, setUserCount] = useState(0);
  const [roomFound, setRoomFound] = useState<boolean>(true);
  const [roomName, setRoomName] = useState<string>("Untitled Room");
  const [listening, setListening] = useState<boolean>(false);
  const [thinking, setThinking] = useState<boolean>(false);
  const [isRoomPrivate, setIsRoomPrivate] = useState<boolean>(false);
  const params = useParams();
  const roomCode = params.roomCode; // get from url

  const wsRef = useRef<WebSocket | null>(null);
  const editorRef = useRef<HTMLTextAreaElement>(null);
  const previewRef = useRef<HTMLDivElement>(null);
  const isUpdatingRef = useRef(false);
  const lastContentRef = useRef("");
  const recogRef = useRef<SpeechRecognition | null>(null);
  const updateTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const fetchRoomData = async (roomCode: string) => {
    try {
      const data = await apiService.fetchRoom(roomCode);
      setRoom(data.code);
      setRoomName(data.name);
      setRoomFound(true);
      setIsRoomPrivate(!data.isPublic);
    } catch {
      setRoom("");
      setRoomFound(false);
    }
  };

  const fetchDocumentsData = async (roomCode: string) => {
    try {
      const documents = await apiService.fetchDocuments(roomCode);
      setDocuments(documents);

      if (!activeDocId && documents && documents.length > 0) {
        setActiveDocId(documents[0].id);
      }
    } catch (error) {
      console.error("error fetching documents:", error);
    }
  };

  const fetchPdfsData = async (roomCode: string) => {
    try {
      const pdfs = await apiService.fetchPdfs(roomCode);
      setPdfs(pdfs);
    } catch (error) {
      console.error("error fetching PDFs:", error);
    }
  };

  const handleCreateDocument = async (title: string | null) => {
    if (!title || !roomCode) return;
    try {
      const data = await apiService.createDocument(roomCode as string, title);
      await fetchDocumentsData(roomCode as string);
      setActiveDocId(data.id);
      setSidebarCollapsed(true);
    } catch (error) {
      let errorMsg = "";
      if (error instanceof Error) {
        errorMsg = ": " + error.message;
      }
      toast.error(`failed to create document:${errorMsg}`, {
        closeButton: true,
      });
    }
  };

  const handleSelectDocument = (docId: number) => {
    setActiveDocId(docId);
    setActivePdfId(null);
    setSidebarCollapsed(true);
  };

  const talkToAi = () => {
    if (listening) {
      recogRef.current?.stop();
      setListening(false);
      setThinking(false);
    } else {
      setListening(true);
      recogRef.current?.start();
    }
  };

  useEffect(() => {
    if (typeof window != "undefined") {
      // CITATION
      // this code is AI assisted
      // Propmt: nothing works on safari. ReferenceError: Can't find variable: SpeechRecognition. altho it works on chrome
      const SpeechRecognitionAPI =
        window.SpeechRecognition || window.webkitSpeechRecognition;
      if (SpeechRecognitionAPI) {
        // END CITATION
        const recognition = new SpeechRecognitionAPI();
        recognition.continuous = false;
        recognition.lang = "en-US";
        recognition.interimResults = false;
        recognition.maxAlternatives = 1;
        recogRef.current = recognition;
      }
    }
    fetchRoomData(roomCode as string);
  }, []);

  useEffect(() => {
    if (room) {
      fetchDocumentsData(room);
      fetchPdfsData(room);
    }
  }, [room]);

  const updatePreview = (text: string) => {
    if (previewRef.current && text !== undefined && text !== null) {
      previewRef.current.innerHTML = marked.parse(text) as string;
      renderMathInElement(previewRef.current, {
        delimiters: [
          { left: "$$", right: "$$", display: true },
          { left: "$", right: "$", display: false },
          { left: "\\[", right: "\\]", display: true },
          { left: "\\(", right: "\\)", display: false },
        ],
        throwOnError: false,
      });
    }
  };

  const sendUpdate = (newContent: string) => {
    if (!activeDocId) return;
    if (
      !isUpdatingRef.current &&
      wsRef.current &&
      wsRef.current.readyState === WebSocket.OPEN
    ) {
      if (newContent === lastContentRef.current) return;
      const operations = computeOperations(lastContentRef.current, newContent);
      if (operations.length > 0) {
        isUpdatingRef.current = true;
        operations.forEach((operation) => {
          wsRef.current?.send(
            JSON.stringify({
              type: "operation",
              operation: operation,
              documentId: activeDocId,
            }),
          );
        });
        lastContentRef.current = newContent;
        updatePreview(newContent);
        setTimeout(() => {
          isUpdatingRef.current = false;
        }, 50);
      }
    }
  };

  useEffect(() => {
    if (!activeDocId) return;
    let shouldReconnect = true;

    if (recogRef.current) {
      recogRef.current.onresult = (e) => {
        setListening(false);
        setThinking(true);
        const prompt = e.results[0][0].transcript;
        try {
          apiService.askAi(
            prompt,
            activeDocId ?? 0,
            roomCode as string,
            editorRef.current?.selectionStart ?? 0,
          );
        } catch (error) {
          let errorMsg = "";
          if (error instanceof Error) {
            errorMsg = ": " + error.message;
          }
          toast.error(`error asking AI${errorMsg}`, {
            closeButton: true,
          });
        }
      };
    }

    const connect = () => {
      const wsUrl = `${apiService.getWsUrl()}/api/ws?roomCode=${roomCode}&docId=${activeDocId}`;
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus("connected");
        setStatusText("Connected");
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as WSEventData;
          switch (data.type) {
            case "init":
              isUpdatingRef.current = true;
              setContent(data.content ?? "");
              lastContentRef.current = data.content ?? "";
              updatePreview(data.content ?? "");
              isUpdatingRef.current = false;
              setUserCount(data.count ?? 0);
              setStatusText("Connected");
              break;
            case "operation":
              if (!isUpdatingRef.current && editorRef.current) {
                isUpdatingRef.current = true;
                const cursorPosition = editorRef.current.selectionStart;
                const scrollPosition = editorRef.current.scrollTop;
                const operation = data.operation;
                const newContent = applyOperation(
                  editorRef.current.value,
                  operation,
                );
                const newCursor = transformCursor(cursorPosition, operation);
                setContent(newContent);
                lastContentRef.current = newContent;
                updatePreview(newContent);
                if (data.userId == "ai") {
                  setThinking(false);
                }
                setTimeout(() => {
                  if (editorRef.current) {
                    editorRef.current.setSelectionRange(newCursor, newCursor);
                    editorRef.current.scrollTop = scrollPosition;
                  }
                  isUpdatingRef.current = false;
                }, 0);
              }
              break;
            case "clientCount":
              setUserCount(data.count ?? 0);
              break;
            case "documentListUpdate":
              if (roomCode) fetchDocumentsData(roomCode as string);
              break;
          }
        } catch (error) {
          console.error("error processing message:", error);
        }
      };

      ws.onerror = (error) => {
        setStatus("disconnected");
        setStatusText("Error");
      };

      ws.onclose = () => {
        setStatus("disconnected");
        setStatusText("Disconnected");
        if (shouldReconnect) setTimeout(connect, 2000);
      };
    };

    connect();

    return () => {
      shouldReconnect = false;
      if (wsRef.current) wsRef.current.close();
      if (updateTimeoutRef.current) clearTimeout(updateTimeoutRef.current);
    };
  }, [activeDocId]);

  const handleEditorChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value;
    setContent(newContent);
    updatePreview(newContent);
    if (updateTimeoutRef.current) clearTimeout(updateTimeoutRef.current);
    updateTimeoutRef.current = setTimeout(() => sendUpdate(newContent), 100);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Tab") {
      e.preventDefault();
      const target = e.target as HTMLTextAreaElement;
      const start = target.selectionStart;
      const end = target.selectionEnd;
      const value = target.value;
      const newContent =
        value.substring(0, start) + "  " + value.substring(end);
      setContent(newContent);
      setTimeout(() => {
        if (editorRef.current) {
          editorRef.current.selectionStart = editorRef.current.selectionEnd =
            start + 2;
        }
      }, 0);
    }
  };

  const activeDoc = documents.find((d) => d.id === activeDocId);
  const activePdf = pdfs.find((p) => p.id === activePdfId);

  return !roomFound ? (
    <div>
      <h1>Room not found</h1>
    </div>
  ) : (
    <div className="flex h-screen overflow-hidden fixed inset-0">
      <Sidebar
        documents={documents}
        activeDocId={activeDocId}
        onSelectDocument={handleSelectDocument}
        onCreateDocument={handleCreateDocument}
        isCollapsed={sidebarCollapsed}
        onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
        roomCode={roomCode as string}
        onUploadPdf={() => fetchPdfsData(roomCode as string)}
        pdfs={pdfs}
        activePdfId={activePdfId}
        onSelectPdf={(id) => {
          setActivePdfId(id);
          setActiveDocId(null);
        }}
      />

      <div className="flex flex-col flex-1 h-screen bg-white overflow-hidden">
        <EditorHeader
          heading={roomName}
          isPrivate={isRoomPrivate}
          documentName={
            activePdfId
              ? activePdf?.filename || ""
              : activeDoc
                ? (activeDoc.title as string)
                : ""
          }
          roomCode={roomCode as string}
          status={statusText}
          numUsers={userCount}
          isConnected={status == "connected"}
          isPdfViewing={!!activePdfId}
        />

        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              className={
                (activePdfId ? "hidden " : "") +
                "shadow fixed top-36 right-10 z-50 text-black font-semibold rounded-full p-4 " +
                (thinking
                  ? "bg-yellow-500"
                  : listening
                    ? "bg-red-500"
                    : "bg-gradient-to-br from-indigo-500 to-purple-800 hover:bg-gradient-to-br hover:from-purple-400 hover:to-indigo-800")
              }
              onClick={() => talkToAi()}
              disabled={thinking}
            >
              {thinking ? (
                <EllipsisIcon></EllipsisIcon>
              ) : listening ? (
                <Pause className="w-5 h-5 text-black" />
              ) : (
                <Sparkles className="w-5 h-5 text-white" />
              )}
            </button>
          </TooltipTrigger>
          <TooltipContent className="bg-white text-black">
            <TooltipArrow className="fill-white" />
            <p>
              {thinking
                ? "Thinking..."
                : listening
                  ? "Stop listening"
                  : "Ask AI to insert content"}
            </p>
            {!thinking && !listening ? <p> at your cursor position</p> : ""}
          </TooltipContent>
        </Tooltip>

        {activePdfId ? (
          <div className="flex flex-col flex-1 overflow-hidden bg-white">
            <div className="px-5 py-[15px] bg-gray-50 border-b border-gray-200">
              <h2 className="text-base font-semibold text-gray-700">
                PDF Viewer
              </h2>
            </div>
            <div className="flex-1 overflow-hidden">
              {pdfs.find((p) => p.id === activePdfId) && (
                <iframe
                  src={`https://docs.google.com/gview?embedded=true&url=${pdfs.find((p) => p.id === activePdfId)?.github_url}`}
                  className="w-full h-full"
                  title="PDF Viewer"
                />
              )}
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-0 flex-1 overflow-hidden">
            <div className="flex flex-col overflow-hidden border-r border-gray-200">
              <div className="px-5 py-[15px] bg-gray-50 border-b border-gray-200">
                <h2 className="text-base font-semibold text-gray-700">
                  Markdown Editor
                </h2>
              </div>
              <MarkdownEditor
                ref={editorRef}
                content={content}
                onChange={handleEditorChange}
                onKeyDown={handleKeyDown}
              />
            </div>
            <div className="flex flex-col overflow-hidden bg-zinc-50">
              <div className="px-5 py-[15px] bg-gray-50 border-b border-gray-200">
                <h2 className="text-base font-semibold text-gray-700">
                  Live Preview
                </h2>
              </div>
              <MarkdownPreview ref={previewRef} />
            </div>
          </div>
        )}

        <footer className="bg-gray-50 px-8 py-4 text-center border-t border-gray-200 text-sm text-gray-500">
          <p>Collaborate in real time</p>
        </footer>
      </div>
    </div>
  );
}
