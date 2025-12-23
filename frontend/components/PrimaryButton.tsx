interface PrimaryButtonProps {
  buttonType: "submit" | "reset" | "button" | undefined;
  label: string;
  isDisabled: boolean;
  className?: string;
}

export function PrimaryButton({
  buttonType,
  label,
  isDisabled,
  className = "",
}: PrimaryButtonProps) {
  return (
    <button
      type={buttonType}
      className={
        "bg-gradient-to-br from-indigo-500 to-purple-800 text-white font-semibold px-4 py-2 rounded " +
        className
      }
      disabled={isDisabled}
    >
      {label}
    </button>
  );
}
