"use client"

import type React from "react"
import { useState, useRef, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Send } from "lucide-react"

// Helper function to generate random pastel colors
const getRandomColor = () => {
  const hue = Math.floor(Math.random() * 360)
  return `hsl(${hue}, 70%, 80%)`
}

type Message = {
  id: number
  username: string
  content: string
  color: string
}

const initialMessages: Message[] = [
  { id: 1, username: "Alice", content: "Hey everyone!", color: getRandomColor() },
  { id: 2, username: "Bob", content: "Hi Alice, how's it going?", color: getRandomColor() },
  { id: 3, username: "Charlie", content: "Hello folks!", color: getRandomColor() },
]

export default function ChatUI() {
  const [messages, setMessages] = useState<Message[]>(initialMessages)
  const [inputMessage, setInputMessage] = useState("")
  const chatContainerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (chatContainerRef.current) {
      chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight
    }
  }, [chatContainerRef.current]) //Corrected dependency

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault()
    if (inputMessage.trim()) {
      const newMessage: Message = {
        id: messages.length + 1,
        username: "You",
        content: inputMessage.trim(),
        color: getRandomColor(),
      }
      setMessages([...messages, newMessage])
      setInputMessage("")
    }
  }

  return (
    <div className="w-full h-full    flex flex-col my-2">
      <div
        ref={chatContainerRef}
        className="flex-1 overflow-y-auto mb-4 border border-gray-200 rounded-md p-4 bg-gray-50"
      >
        {messages.map((message) => (
          <div key={message.id} className="mb-3">
            <span style={{ color: message.color }} className="font-semibold mr-2">
              {message.username}:
            </span>
            <span className="text-gray-800">{message.content}</span>
          </div>
        ))}
      </div>

      {/* Message input form */}
      <form onSubmit={handleSendMessage} className="flex gap-2">
        <Input
          type="text"
          placeholder="Type a message..."
          value={inputMessage}
          onChange={(e) => setInputMessage(e.target.value)}
          className="flex-1"
        />
        <Button type="submit" className="bg-purple-600 hover:bg-purple-700 text-white">
          <Send className="w-4 h-4" />
        </Button>
      </form>
    </div>
  )
}

