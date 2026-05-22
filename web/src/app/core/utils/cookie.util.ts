export function getCookie(name: string): string | null {
  if (typeof document === 'undefined' || !document.cookie) {
    return null;
  }

  const encodedName = encodeURIComponent(name);
  const parts = document.cookie.split(';');

  for (const part of parts) {
    const trimmed = part.trim();
    if (trimmed.startsWith(`${encodedName}=`)) {
      return decodeURIComponent(trimmed.substring(encodedName.length + 1));
    }
  }

  return null;
}
