interface TextInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  className?: string;
  type?: React.HTMLInputTypeAttribute;
}

export function TextInput({
  value,
  onChange,
  placeholder,
  className = "",
  type = "text",
}: TextInputProps) {
  return (
    <input
      type={type}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={
        "flex-1 px-4 py-2 rounded border border-indigo-200 focus:outline-none focus:ring-2 focus:ring-indigo-300 " +
        className
      }
      aria-label="New room name"
    />
  );
}
