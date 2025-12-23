export function getUsername(): string | null {
  if (typeof document === "undefined") {
    return null;
  }
  return document.cookie.replace(
    /(?:(?:^|.*;\s*)username\s*\=\s*([^;]*).*$)|^.*$/,
    "$1",
  );
}
