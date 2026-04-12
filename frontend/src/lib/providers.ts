import groqIcon from "@/assets/provider-icons/groq.svg";
import openaiCompatIcon from "@/assets/provider-icons/openai-compatible.svg";

export interface ProviderPreset {
  name: string;
  description: string;
  icon: string; // path to SVG asset
  keyPlaceholder: string;
  keyUrl: string;
  keyUrlLabel: string;
  requiresBaseUrl?: boolean; // if true, base URL is shown by default (not behind Advanced)
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
  "openai-compatible": {
    name: "OpenAI-Compatible",
    description:
      "Any service with an OpenAI-compatible API — OpenAI, Ollama, LM Studio, vLLM, Together, Fireworks, and more.",
    icon: openaiCompatIcon,
    keyPlaceholder: "sk-... (optional for local)",
    keyUrl: "",
    keyUrlLabel: "",
    requiresBaseUrl: true,
  },
};
