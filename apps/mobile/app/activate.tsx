import { useActivationLink } from "@/activation/useActivationLinkState";

export default function ActivateRoute() {
  useActivationLink();
  return null;
}
