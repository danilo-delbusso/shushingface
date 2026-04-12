import groqIcon from "@/assets/provider-icons/groq.svg";

export interface ProviderPreset {
  name: string;
  description: string;
  icon: string; // path to SVG asset
  keyPlaceholder: string;
  keyUrl: string;
  keyUrlLabel: string;
}

/** Known provider presets — everything except the API key is pre-configured. */
export const providerPresets: Record<string, ProviderPreset> = {
  groq: {
    name: "Groq",
    description:
      "Ultra-fast inference with Llama, Whisper, Qwen, and more. Free tier available.",
    icon: groqIcon,
    keyPlaceholder: "gsk_...",
    keyUrl: "https://console.groq.com/keys",
    keyUrlLabel: "Get a free key",
  },
};
