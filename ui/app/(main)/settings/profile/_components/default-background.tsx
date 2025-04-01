import React from "react";

export default function DefaultBackgound() {
  return (
    <div className="relative h-40 inset-0 grid grid-cols-6 gap-2 p-2 bg-gray-800 rounded-lg">
      {[...Array(18)].map((_, i) => (
        <svg
          key={i}
          className="w-8 h-8 text-white opacity-25"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path d="M21 3H3v18h18V3zm-9 14H7v-4h5v4zm0-6H7V7h5v4zm6 6h-4v-4h4v4zm0-6h-4V7h4v4z" />
        </svg>
      ))}
    </div>
  );
}
