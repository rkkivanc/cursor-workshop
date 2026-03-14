"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

interface Message {
  role: "user" | "assistant";
  content: string;
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function submitMessage() {
    const text = input.trim();
    if (!text || loading) return;

    const userMessage: Message = { role: "user", content: text };
    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    setInput("");
    setLoading(true);

    const assistantMessage: Message = { role: "assistant", content: "" };
    setMessages([...updatedMessages, assistantMessage]);

    try {
      const res = await fetch("/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ messages: updatedMessages }),
      });

      if (!res.ok || !res.body) {
        throw new Error("Sunucudan yanıt alınamadı");
      }

      const reader = res.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split("\n").filter(Boolean);

        for (const line of lines) {
          try {
            const json = JSON.parse(line);
            if (json.message?.content) {
              setMessages((prev) => {
                const updated = [...prev];
                updated[updated.length - 1] = {
                  role: "assistant",
                  content: updated[updated.length - 1].content + json.message.content,
                };
                return updated;
              });
            }
          } catch {
            // Partial JSON line, skip
          }
        }
      }
    } catch (err) {
      setMessages((prev) => {
        const updated = [...prev];
        updated[updated.length - 1] = {
          role: "assistant",
          content: "Hata: " + (err instanceof Error ? err.message : "Bilinmeyen hata"),
        };
        return updated;
      });
    } finally {
      setLoading(false);
    }
  }

  function handleFormSubmit(e: React.FormEvent) {
    e.preventDefault();
    submitMessage();
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submitMessage();
    }
  }

  return (
    <div className="flex h-screen flex-col bg-zinc-50 dark:bg-zinc-900">
      {/* Header */}
      <header className="flex items-center justify-between border-b border-zinc-200 bg-white px-6 py-4 dark:border-zinc-700 dark:bg-zinc-800">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-green-500 text-sm font-bold text-white">
            AI
          </div>
          <div>
            <h1 className="text-base font-semibold text-zinc-900 dark:text-zinc-50">
              Ollama Sohbet
            </h1>
            <p className="text-xs text-zinc-500 dark:text-zinc-400">gemma3:1b</p>
          </div>
        </div>
        <Link
          href="/"
          className="text-sm text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-50"
        >
          ← Ana Sayfa
        </Link>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-6">
        {messages.length === 0 && (
          <div className="flex h-full flex-col items-center justify-center gap-3 text-center text-zinc-400">
            <div className="text-4xl">💬</div>
            <p className="text-lg font-medium">Sohbete başlayın</p>
            <p className="max-w-sm text-sm">
              gemma3:1b modeli ile sohbet etmek için aşağıya bir mesaj yazın.
            </p>
          </div>
        )}
        <div className="mx-auto max-w-2xl space-y-4">
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
            >
              <div
                className={`max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed ${
                  msg.role === "user"
                    ? "bg-zinc-900 text-white dark:bg-zinc-50 dark:text-zinc-900"
                    : "bg-white text-zinc-800 shadow-sm ring-1 ring-zinc-100 dark:bg-zinc-800 dark:text-zinc-100 dark:ring-zinc-700"
                }`}
              >
                {msg.content || (
                  <span className="inline-flex gap-1 text-zinc-400">
                    <span className="animate-bounce" style={{ animationDelay: "0ms" }}>●</span>
                    <span className="animate-bounce" style={{ animationDelay: "150ms" }}>●</span>
                    <span className="animate-bounce" style={{ animationDelay: "300ms" }}>●</span>
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="border-t border-zinc-200 bg-white px-4 py-4 dark:border-zinc-700 dark:bg-zinc-800">
        <form
          onSubmit={handleFormSubmit}
          className="mx-auto flex max-w-2xl items-end gap-3"
        >
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Mesajınızı yazın... (Enter ile gönder)"
            rows={1}
            disabled={loading}
            className="flex-1 resize-none rounded-xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-sm text-zinc-900 placeholder-zinc-400 focus:border-zinc-400 focus:outline-none focus:ring-0 disabled:opacity-50 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-50 dark:placeholder-zinc-500"
          />
          <button
            type="submit"
            disabled={loading || !input.trim()}
            className="flex h-11 w-11 items-center justify-center rounded-xl bg-zinc-900 text-white transition-colors hover:bg-zinc-700 disabled:opacity-40 dark:bg-zinc-50 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="currentColor"
              className="h-5 w-5"
            >
              <path d="M3.478 2.405a.75.75 0 0 0-.926.94l2.432 7.905H13.5a.75.75 0 0 1 0 1.5H4.984l-2.432 7.905a.75.75 0 0 0 .926.94 60.519 60.519 0 0 0 18.445-8.986.75.75 0 0 0 0-1.218A60.517 60.517 0 0 0 3.478 2.405Z" />
            </svg>
          </button>
        </form>
      </div>
    </div>
  );
}
