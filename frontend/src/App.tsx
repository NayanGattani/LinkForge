import { FormEvent, useState } from "react";
import { BackgroundShapes } from "./BackgroundShapes";
import { useLocalStorage } from "./hooks/useLocalStorage";

interface HistoryItem {
  id: string;
  originalUrl: string;
  shortUrl: string;
  alias?: string;
  expiresAt?: string;
}

function CopyIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24">
      <path d="M9 9h10v10H9z" fill="currentColor" stroke="black" strokeWidth="1.5" />
      <path d="M5 5h10v2H7v8H5z" fill="currentColor" stroke="black" strokeWidth="1.5" />
    </svg>
  );
}

function HistoryCard({ item, onDelete }: { item: HistoryItem; onDelete: (id: string) => void }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(item.shortUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="flex flex-col gap-3 border-4 border-black bg-[#ffb8eb] p-4 shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0 flex-1 overflow-hidden">
        <p className="truncate text-sm font-bold opacity-75" title={item.originalUrl}>{item.originalUrl}</p>
        <a className="mt-1 block truncate text-lg font-black underline decoration-2 underline-offset-4" href={item.shortUrl} target="_blank" rel="noreferrer">
          {item.shortUrl.replace(/^https?:\/\//, "")}
        </a>
        {item.expiresAt && <p className="mt-1 text-xs font-black uppercase">Expires: {new Date(item.expiresAt).toLocaleString()}</p>}
      </div>
      <div className="flex shrink-0 gap-2">
        <button className="inline-flex items-center gap-1 border-4 border-black bg-[#ffe45e] px-3 py-2 text-xs font-black uppercase shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]" onClick={copy} type="button">
          <CopyIcon /> {copied ? "Copied" : "Copy"}
        </button>
        <button className="border-4 border-black bg-[#ff5e5e] px-3 py-2 text-xs font-black uppercase shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]" onClick={() => onDelete(item.id)} type="button">Delete</button>
      </div>
    </div>
  );
}

function App() {
  const [url, setUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [expiry, setExpiry] = useState("0");
  const [shortUrl, setShortUrl] = useState("");
  const [expiresAt, setExpiresAt] = useState<string | undefined>();
  const [isCopied, setIsCopied] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [history, setHistory] = useLocalStorage<HistoryItem[]>("linkforge-history", []);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedUrl = url.trim();
    if (!trimmedUrl) return;
    setIsLoading(true);
    setError("");

    try {
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/shorten`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: trimmedUrl, alias: alias.trim(), expires_in: Number(expiry) }),
      });
      const data = await response.json();
      if (!response.ok) {
        setError(data.error || "Failed to shorten URL. Please try again.");
        return;
      }
      setShortUrl(data.short_url);
      setExpiresAt(data.expires_at);
      setIsCopied(false);
      setHistory(prev => [{ id: data.id, originalUrl: trimmedUrl, shortUrl: data.short_url, alias: alias.trim() || undefined, expiresAt: data.expires_at }, ...prev.filter(i => i.originalUrl !== trimmedUrl)].slice(0, 50));
    } catch {
      setError("Could not reach the LinkForge API.");
    } finally {
      setIsLoading(false);
    }
  };

  const copy = async () => {
    if (!shortUrl) return;
    await navigator.clipboard.writeText(shortUrl);
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 1500);
  };

  return (
    <main className="relative z-0 flex min-h-screen items-center justify-center px-5 pb-24 pt-10 text-black">
      <BackgroundShapes />
      <button className="absolute right-6 top-6 z-20 border-4 border-black bg-[#ff9900] px-4 py-2 font-black uppercase shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]" onClick={() => setIsModalOpen(true)} type="button">History</button>

      <section className="relative z-10 w-full max-w-4xl">
        <div className="border-4 border-black bg-white p-6 shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] sm:p-10">
          <p className="mb-4 inline-block border-4 border-black bg-[#7cff65] px-3 py-1 text-sm font-black uppercase tracking-[0.2em] shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">Fast URL shortener</p>
          <h1 className="text-5xl font-black uppercase leading-none text-[#e5e5e5] [-webkit-text-stroke:2px_black] [text-shadow:4px_4px_0_#000000] sm:text-7xl">LinkForge</h1>
          <p className="mt-5 max-w-2xl text-base font-bold uppercase tracking-[0.08em] sm:text-lg">Create compact links with custom aliases, automatic expiration, and built-in redirect analytics.</p>

          <form className="mt-8 space-y-4" onSubmit={handleSubmit}>
            <div className="flex flex-col gap-4 sm:flex-row">
              <input className="min-w-0 flex-1 border-4 border-black bg-white px-5 py-4 font-bold outline-none shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]" onChange={e => setUrl(e.target.value)} placeholder="https://your-long-url.com/..." type="url" value={url} disabled={isLoading} required />
              <button className="border-4 border-black bg-[#00a6ff] px-8 py-4 text-lg font-black uppercase shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]" disabled={isLoading} type="submit">{isLoading ? "Working..." : "Shorten"}</button>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <input className="border-4 border-black bg-white px-4 py-3 font-bold shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]" onChange={e => setAlias(e.target.value)} placeholder="Custom alias (optional)" value={alias} disabled={isLoading} maxLength={32} pattern="[A-Za-z0-9_-]*" />
              <select className="border-4 border-black bg-white px-4 py-3 font-bold shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]" value={expiry} onChange={e => setExpiry(e.target.value)} disabled={isLoading}>
                <option value="0">No expiration</option>
                <option value="3600">1 hour</option>
                <option value="86400">24 hours</option>
                <option value="604800">7 days</option>
                <option value="2592000">30 days</option>
              </select>
            </div>
          </form>

          {error && <div className="mt-8 border-4 border-black bg-[#ff5e5e] p-4 text-center font-bold uppercase shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]">{error}</div>}
          {shortUrl && !error && (
            <div className="mt-8 border-4 border-black bg-[#a4f58c] p-5 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
              <p className="text-sm font-black uppercase tracking-[0.18em]">Your short link</p>
              <div className="mt-2 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <a className="break-all text-2xl font-black underline decoration-4 underline-offset-4" href={shortUrl}>{shortUrl.replace(/^https?:\/\//, "")}</a>
                <button className="inline-flex items-center gap-2 border-4 border-black bg-[#ffe45e] px-4 py-3 text-sm font-black uppercase shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]" onClick={copy} type="button"><CopyIcon />{isCopied ? "Copied" : "Copy link"}</button>
              </div>
              {expiresAt && <p className="mt-3 text-xs font-black uppercase">Expires {new Date(expiresAt).toLocaleString()}</p>}
            </div>
          )}
        </div>
      </section>

      <footer className="absolute bottom-6 left-6 z-10 font-bold uppercase text-black sm:bottom-8 sm:left-8">
        <p className="inline-block border-2 border-black bg-[#ff9900] px-2 py-1 text-xs tracking-widest shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] sm:text-sm">LinkForge</p>
        <p className="mt-2 text-xs tracking-wider">Go + React + PostgreSQL + Redis</p>
      </footer>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm">
          <div className="flex w-full max-w-3xl flex-col border-4 border-black bg-white shadow-[10px_10px_0px_0px_rgba(0,0,0,1)]">
            <div className="flex items-center justify-between border-b-4 border-black bg-[#ff9900] p-4"><h2 className="text-2xl font-black uppercase">Recent links</h2><button className="border-4 border-black bg-[#ff5e5e] px-3 py-1 font-black" onClick={() => setIsModalOpen(false)} type="button">Close</button></div>
            <div className="flex max-h-[70vh] flex-col gap-4 overflow-y-auto bg-[#f4f4f4] p-4 sm:p-6">
              {history.length ? history.map(item => <HistoryCard item={item} key={item.id} onDelete={id => setHistory(prev => prev.filter(i => i.id !== id))} />) : <p className="py-8 text-center text-lg font-bold uppercase text-black/50">No links created yet.</p>}
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

export default App;
