import type { RoomResponse } from "../types/index";

export const getApiUrl = () => {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  if (typeof window !== "undefined") {
    return "";
  }
  return "http://localhost:3000";
};

export const getWsUrl = () => {
  if (process.env.NEXT_PUBLIC_API_URL) {
    const apiUrl = new URL(process.env.NEXT_PUBLIC_API_URL);
    const protocol = apiUrl.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${apiUrl.host}`;
  }
  if (typeof window !== "undefined") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}`;
  }
  return "ws://localhost:3000";
};

export const signInWithGithub = async (code: string) => {
  const response = await fetch(`${getApiUrl()}/api/auth/github`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });

  return response;
};

export const signUp = async (username: string, password: string) => {
  const response = await fetch(`${getApiUrl()}/api/auth/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: username, password: password }),
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });

  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
};

export const signIn = async (username: string, password: string) => {
  const response = await fetch(`${getApiUrl()}/api/auth/signin`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });

  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
};

// Room API
export const fetchRooms = async (limit: number, offset: number) => {
  const response = await fetch(
    `${getApiUrl()}/api/rooms?limit=${limit}&offset=${offset}`,
    {
      credentials:
        process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
    },
  );
  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
  const data = await response.json();
  let hasMore = false;
  if (data.hasMoreData) hasMore = true;
  return { rooms: data.rooms || [], hasMore: hasMore };
};

export const askAi = async (
  prompt: string,
  docId: number,
  roomCode: string,
  cursorPosition: number,
) => {
  const response = await fetch(`${getApiUrl()}/api/ai`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      documentId: docId,
      prompt: prompt,
      roomCode: roomCode,
      cursorPosition: cursorPosition,
    }),
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });

  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
};

export const fetchRoom = async (roomCode: string): Promise<RoomResponse> => {
  const response = await fetch(`${getApiUrl()}/api/rooms/${roomCode}`, {
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });
  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
  const data = await response.json();
  return data;
};

export const createRoom = async (name: string, isPublic: boolean) => {
  const response = await fetch(`${getApiUrl()}/api/rooms`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, isPublic }),
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });
  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
  const data = await response.json();
  return data;
};

// Document API
export const fetchDocuments = async (roomCode: string) => {
  const response = await fetch(
    `${getApiUrl()}/api/documents?roomCode=${roomCode}`,
    {
      credentials:
        process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
    },
  );
  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
  const data = await response.json();
  return data.documents || [];
};

export const createDocument = async (roomCode: string, title: string) => {
  const response = await fetch(
    `${getApiUrl()}/api/documents?roomCode=${roomCode}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title }),
      credentials:
        process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
    },
  );
  if (!response.ok) {
    const errorMessage = await response.text();
    throw new Error(errorMessage);
  }
  const data = await response.json();
  return data;
};

export const fetchPdfs = async (roomCode: string) => {
  const response = await fetch(`${getApiUrl()}/api/pdfs?roomCode=${roomCode}`, {
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });
  if (!response.ok) {
    throw new Error("Failed to fetch PDFs");
  }
  const data = await response.json();
  return data.pdfs || [];
};

export const uploadPdf = async (roomCode: string, file: File) => {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch(
    `${getApiUrl()}/api/pdfs/upload?roomCode=${roomCode}`,
    {
      method: "POST",
      body: formData,
      credentials:
        process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
    },
  );
  if (!response.ok) {
    let errorMessage = "Failed to upload PDF";
    try {
      const errorText = await response.text();
      try {
        const errorData = JSON.parse(errorText);
        errorMessage =
          errorData.message || errorData || errorText || "Failed to upload PDF";
      } catch {
        errorMessage =
          errorText || `Upload failed with status ${response.status}`;
      }
    } catch (e) {
      errorMessage = `Upload failed with status ${response.status}`;
    }
    const error = new Error(errorMessage);
    (error as any).status = response.status;
    throw error;
  }
  const data = await response.json();
  return data;
};

export const deletePdf = async (pdfId: number) => {
  const response = await fetch(`${getApiUrl()}/api/pdfs?pdfId=${pdfId}`, {
    method: "DELETE",
    credentials:
      process.env.NEXT_PUBLIC_ENV == "production" ? "same-origin" : "include",
  });
  if (!response.ok) {
    let errorMessage = "Failed to delete PDF";
    try {
      const errorText = await response.text();
      try {
        const errorData = JSON.parse(errorText);
        errorMessage =
          errorData.message || errorData || errorText || "Failed to delete PDF";
      } catch {
        errorMessage =
          errorText || `Delete failed with status ${response.status}`;
      }
    } catch (e) {
      errorMessage = `Delete failed with status ${response.status}`;
    }
    const error = new Error(errorMessage);
    (error as any).status = response.status;
    throw error;
  }
  const data = await response.json();
  return data;
};
