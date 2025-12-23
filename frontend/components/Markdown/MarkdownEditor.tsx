import React from "react";

interface MarkdownEditorProps {
  content: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLTextAreaElement>) => void;
}

export const MarkdownEditor = React.forwardRef<
  HTMLTextAreaElement,
  MarkdownEditorProps
>(({ content, onChange, onKeyDown }, ref) => {
  return (
    <textarea
      ref={ref}
      id="markdown-editor"
      placeholder="Start typing your markdown here..."
      value={content}
      onChange={onChange}
      onKeyDown={onKeyDown}
      className="flex-1 p-5 border-0 outline-none font-mono text-sm leading-relaxed resize-none bg-white text-gray-800 placeholder:text-gray-400"
    />
  );
});

MarkdownEditor.displayName = "MarkdownEditor";
