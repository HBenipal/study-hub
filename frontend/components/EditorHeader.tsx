import Link from "next/link";

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { toast } from "sonner";
import { TooltipArrow } from "@radix-ui/react-tooltip";

interface EditorHeaderProps {
  heading: string;
  documentName: string;
  roomCode: string;
  status: string;
  numUsers: number;
  isConnected: boolean;
  isPrivate: boolean;
  isPdfViewing?: boolean;
}

const copyCode = (code: string) => {
  navigator.clipboard.writeText(code);
  toast.success("Copied code to clipboard", {
    closeButton: true,
  });
};

export function EditorHeader({
  heading,
  documentName,
  roomCode,
  status,
  numUsers,
  isConnected,
  isPrivate,
  isPdfViewing,
}: EditorHeaderProps) {
  return (
    <header className="bg-gradient-to-br from-indigo-500 to-purple-800 text-white px-7 py-5 flex justify-between items-center shadow-md">
      <div>
        <h1 className="text-2xl font-semibold m-0 mb-2">{heading}</h1>
        <div className="flex gap-4">
          <span className="bg-white/15 px-3 py-1 rounded-md">
            <div className="font-bold text-white">{documentName}</div>
          </span>
          <Tooltip>
            <TooltipTrigger asChild>
              <div
                className="bg-white/15 px-3 py-1 rounded-md cursor-pointer flex items-center justify-center gap-1"
                onClick={() => copyCode(roomCode)}
              >
                <div className="max-sm:hidden">Code: </div>
                <div className="font-bold text-white">{roomCode}</div>
              </div>
            </TooltipTrigger>
            <TooltipContent className="bg-white text-black">
              <TooltipArrow className="fill-white" />
              <p>Click to copy code</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
      <div className="flex items-center gap-4">
        {!isPdfViewing && (
          <div className="flex items-center justify-around gap-3 max-md:hidden">
            <span
              className={`w-2.5 h-2.5 animate-pulse rounded-full inline-block ${isConnected ? "bg-green-400" : "bg-red-500"}`}
            ></span>
            <span className="text-sm">{status}</span>
          </div>
        )}
        <div className="flex justify-center items-center max-sm:flex-col gap-4">
          <span className="bg-white/20 px-3 py-1 rounded-full text-sm font-medium">
            {numUsers} {numUsers === 1 ? "user" : "users"} online
          </span>
          <Tooltip>
            <TooltipTrigger asChild>
              <Link href={"/"}>
                <span className="bg-gradient-to-tr from-red-500 to-red-800 text-white px-3 py-1.5 rounded-full text-sm font-medium">
                  Leave room
                </span>
              </Link>
            </TooltipTrigger>
            <TooltipContent className="bg-white text-black">
              <TooltipArrow className="fill-white" />
              {isPrivate ? (
                <p className="text-red-500">
                  This room is private! Copy link before leaving
                </p>
              ) : (
                <p className="text-green-500">
                  This room is public! You can access it again from the home
                  page
                </p>
              )}
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </header>
  );
}
