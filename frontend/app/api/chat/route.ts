import { NextRequest } from "next/server";

const OLLAMA_URL = process.env.OLLAMA_URL ?? "http://localhost:11434";

export async function POST(request: NextRequest) {
  const body = await request.json();

  let ollamaResponse: Response;
  try {
    ollamaResponse = await fetch(`${OLLAMA_URL}/api/chat`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        model: body.model ?? "gemma3:1b",
        messages: body.messages,
        stream: true,
      }),
    });
  } catch {
    return new Response(
      JSON.stringify({ error: "Ollama servisine bağlanılamadı. Servisin çalıştığından emin olun." }),
      { status: 503, headers: { "Content-Type": "application/json" } }
    );
  }

  if (!ollamaResponse.ok) {
    return new Response(
      JSON.stringify({ error: "Ollama servisi yanıt vermedi" }),
      { status: ollamaResponse.status, headers: { "Content-Type": "application/json" } }
    );
  }

  // Stream the response directly from Ollama to the client
  return new Response(ollamaResponse.body, {
    headers: { "Content-Type": "application/x-ndjson" },
  });
}
