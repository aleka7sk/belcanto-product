/**
 * Hosts that are safe to use over plain HTTP in an explicitly non-production
 * build. This covers simulators, Android's host alias and private LAN devices
 * without allowing a development flag to downgrade an arbitrary public host.
 */
function normalizedHostname(hostname: string): string {
  return hostname
    .trim()
    .toLowerCase()
    .replace(/^\[|\]$/g, "");
}

export function isLoopbackHost(hostname: string): boolean {
  const normalized = normalizedHostname(hostname);
  if (normalized === "localhost" || normalized === "::1") return true;
  const octets = normalized.split(".").map((part) => Number(part));
  return (
    octets.length === 4 &&
    octets.every(
      (octet, index) =>
        Number.isInteger(octet) &&
        octet >= 0 &&
        octet <= 255 &&
        String(octet) === normalized.split(".")[index],
    ) &&
    octets[0] === 127
  );
}

export function isPrivateDevelopmentHost(hostname: string): boolean {
  const normalized = normalizedHostname(hostname);

  if (isLoopbackHost(normalized)) return true;
  if (/^(?:fc|fd)[0-9a-f]{2}:/i.test(normalized)) return true;

  const octets = normalized.split(".").map((part) => Number(part));
  if (
    octets.length !== 4 ||
    octets.some(
      (octet, index) =>
        !Number.isInteger(octet) ||
        octet < 0 ||
        octet > 255 ||
        String(octet) !== normalized.split(".")[index],
    )
  ) {
    return false;
  }

  const [first, second] = octets;
  return (
    first === 10 ||
    (first === 172 && second !== undefined && second >= 16 && second <= 31) ||
    (first === 192 && second === 168)
  );
}
