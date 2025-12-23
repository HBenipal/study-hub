import Navbar from "@/components/Navbar";
import React from "react";

export default function Home() {
  return (
    <div className="min-h-screen bg-zinc-50 p-8">
      <div className="max-w-6xl mx-auto">
        <div className="mb-12">
          <Navbar />
        </div>
        <ul className="list-disc m-10">
          <li className="list-item">
            Icons from{" "}
            <a className="text-blue-700 underline" href="https://lucide.dev">
              Lucide react
            </a>
            , which is under the{" "}
            <a
              className="text-blue-700 underline"
              href="https://lucide.dev/license"
            >
              Lucide License
            </a>
          </li>
          <li className="list-item">
            Tooltip and Toast components from{" "}
            <a
              className="text-blue-700 underline"
              href="https://github.com/shadcn-ui/ui?tab=MIT-1-ov-file"
            >
              shadcn-ui
            </a>
            , which is under the{" "}
            <a
              className="text-blue-700 underline"
              href="https://github.com/shadcn-ui/ui/blob/main/LICENSE.md"
            >
              MIT license
            </a>
          </li>
          <li className="list-item">
            The Roboto font is from{" "}
            <a
              className="text-blue-700 underline"
              href="https://fonts.google.com/specimen/Roboto"
            >
              Google Fonts
            </a>
            , under the{" "}
            <a
              className="text-blue-700 underline"
              href="https://fonts.google.com/specimen/Roboto/license"
            >
              SIL Open Font License
            </a>
          </li>
        </ul>
      </div>
    </div>
  );
}
