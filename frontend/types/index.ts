export interface Room {
  code: string;
  name: string;
  activeUsers: number;
  lastUpdated: string;
}

export interface DocumentItem {
  id: number;
  title: string;
}

export interface SidebarProps {
  documents: DocumentItem[];
  activeDocId: number | null;
  onSelectDocument: (id: number) => void;
  onCreateDocument: (title: string) => void;
  isCollapsed: boolean;
  onToggle: () => void;
}

export interface WSEventData {
  type: "init" | "operation" | "clientCount" | "documentListUpdate" | "error";
  userName: string | null;
  userId: string | null;
  version: number | null;
  content: string | null;
  count: number | null;
  operation: "insert" | "delete" | null;
  message: string | null;
}

export interface RoomResponse {
  name: string;
  code: string;
  isPublic: boolean;
}
