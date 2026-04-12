/** Known provider presets — everything except the API key is pre-configured. */
export const providerPresets: Record<
  string,
  {
    name: string;
    description: string;
    keyPlaceholder: string;
    keyUrl: string;
    keyUrlLabel: string;
  }
> = {
  groq: {
    name: "Groq",
    description:
      "Ultra-fast inference with Llama, Whisper, Qwen, and more. Free tier available.",
    keyPlaceholder: "gsk_...",
    keyUrl: "https://console.groq.com/keys",
    keyUrlLabel: "Get a free key",
  },
};
