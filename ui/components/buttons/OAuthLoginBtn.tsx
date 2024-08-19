"use client";

export default function OAuthLoginBtn({ url, children }: { url: string, children: React.ReactNode }) {
  return (
    <button
      onClick={() => {
        window.location.href = url;
      }}
      className="flex items-center justify-center gap-1 rounded-md border-gray-300 border w-[200px] h-[50px] hover:cursor-pointer"
    >
        {children}
    </button>
  );
}
