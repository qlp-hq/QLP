'use client'

import React, { useState, useRef, useEffect } from 'react'
import { MessageCircle, X, Send, Bot, User, Sparkles, Loader2 } from 'lucide-react'

const suggestions = [
  "How does LLOM work?",
  "What's included in the generated code?",
  "Can I customize the AI agents?",
  "Show me pricing options"
]

const initialMessages = [
  {
    id: 1,
    type: 'bot',
    content: "Hi! I'm Quantum, your AI assistant. I can help you understand how QuantumLayer works and answer any questions about building microservices with AI. What would you like to know?",
    timestamp: new Date()
  }
]

export default function AIAssistant() {
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState(initialMessages)
  const [input, setInput] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSend = (text?: string) => {
    const messageText = text || input.trim()
    if (!messageText) return

    // Add user message
    const userMessage = {
      id: messages.length + 1,
      type: 'user',
      content: messageText,
      timestamp: new Date()
    }
    
    setMessages(prev => [...prev, userMessage])
    setInput('')
    setIsTyping(true)

    // Simulate bot response
    setTimeout(() => {
      const botMessage = {
        id: messages.length + 2,
        type: 'bot',
        content: getBotResponse(messageText),
        timestamp: new Date()
      }
      setMessages(prev => [...prev, botMessage])
      setIsTyping(false)
    }, 1500)
  }

  const getBotResponse = (message: string) => {
    const lowerMessage = message.toLowerCase()
    
    if (lowerMessage.includes('llom') || lowerMessage.includes('how') && lowerMessage.includes('work')) {
      return "LLOM (Large Language Orchestration Model) is our revolutionary approach to code generation. Instead of using templates, we dynamically orchestrate 6-16 specialized AI agents that collaborate to build your unique application. Each agent is an expert in their domain - frontend, backend, security, databases, etc. They work together to create production-ready microservices in just 12.5 seconds!"
    }
    
    if (lowerMessage.includes('price') || lowerMessage.includes('cost')) {
      return "We offer three pricing tiers:\n\n• **Starter (Free)**: 3 projects, 1,000 API calls/month\n• **Professional ($79/month)**: Unlimited projects, 100,000 API calls\n• **Enterprise (Custom)**: Dedicated support, on-premise options, custom SLAs\n\nAll plans include full access to our AI agents and production-ready code generation!"
    }
    
    if (lowerMessage.includes('custom') || lowerMessage.includes('agent')) {
      return "Yes! With QuantumLayer Enterprise, you can train custom AI agents specific to your tech stack and coding standards. Our agents learn from your codebase and can follow your architectural patterns, ensuring generated code matches your organization's best practices."
    }
    
    if (lowerMessage.includes('include') || lowerMessage.includes('what') && lowerMessage.includes('get')) {
      return "Every QuantumLayer project includes:\n\n✓ Complete source code (no lock-in)\n✓ Comprehensive test suites\n✓ API documentation\n✓ Docker containers\n✓ CI/CD pipelines\n✓ Database migrations\n✓ Security configurations\n✓ Monitoring setup\n\nEverything you need to go from idea to production!"
    }
    
    return "That's a great question! QuantumLayer uses AI to transform your ideas into production-ready microservices in seconds. Would you like to know more about our LLOM architecture, pricing, or see a demo?"
  }

  return (
    <>
      {/* Chat Button */}
      <button
        onClick={() => setIsOpen(true)}
        className={`fixed bottom-6 right-6 p-4 rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 shadow-2xl hover:scale-105 transition-all duration-300 z-40 ${
          isOpen ? 'opacity-0 pointer-events-none' : ''
        }`}
      >
        <div className="relative">
          <MessageCircle className="w-6 h-6 text-white" />
          <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-400 rounded-full animate-pulse" />
        </div>
      </button>

      {/* Chat Window */}
      <div className={`fixed bottom-6 right-6 w-96 h-[600px] glass rounded-2xl shadow-2xl border border-white/10 flex flex-col z-50 transition-all duration-300 transform ${
        isOpen ? 'scale-100 opacity-100' : 'scale-95 opacity-0 pointer-events-none'
      }`}>
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-white/10">
          <div className="flex items-center gap-3">
            <div className="relative">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                <Bot className="w-6 h-6 text-white" />
              </div>
              <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-400 rounded-full border-2 border-black" />
            </div>
            <div>
              <h3 className="font-semibold">Quantum Assistant</h3>
              <p className="text-xs text-gray-400">Always here to help</p>
            </div>
          </div>
          <button
            onClick={() => setIsOpen(false)}
            className="p-2 rounded-lg hover:bg-white/10 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div className={`flex items-start gap-3 max-w-[80%] ${
                message.type === 'user' ? 'flex-row-reverse' : ''
              }`}>
                <div className={`p-2 rounded-lg ${
                  message.type === 'user'
                    ? 'bg-gradient-to-r from-blue-600 to-purple-600'
                    : 'bg-white/10'
                }`}>
                  {message.type === 'user' ? (
                    <User className="w-4 h-4 text-white" />
                  ) : (
                    <Bot className="w-4 h-4 text-white" />
                  )}
                </div>
                <div className={`px-4 py-2 rounded-2xl ${
                  message.type === 'user'
                    ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white'
                    : 'glass border border-white/10'
                }`}>
                  <p className="text-sm whitespace-pre-wrap">{message.content}</p>
                </div>
              </div>
            </div>
          ))}
          
          {isTyping && (
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-lg bg-white/10">
                <Bot className="w-4 h-4 text-white" />
              </div>
              <div className="px-4 py-2 rounded-2xl glass border border-white/10">
                <div className="flex gap-1">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" />
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce delay-100" />
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce delay-200" />
                </div>
              </div>
            </div>
          )}
          
          <div ref={messagesEndRef} />
        </div>

        {/* Suggestions */}
        {messages.length === 1 && (
          <div className="px-4 pb-2">
            <p className="text-xs text-gray-400 mb-2">Quick questions:</p>
            <div className="flex flex-wrap gap-2">
              {suggestions.map((suggestion, index) => (
                <button
                  key={index}
                  onClick={() => handleSend(suggestion)}
                  className="px-3 py-1 text-xs glass border border-white/10 rounded-full hover:border-white/20 transition-all"
                >
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Input */}
        <div className="p-4 border-t border-white/10">
          <form onSubmit={(e) => { e.preventDefault(); handleSend(); }} className="flex gap-2">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ask me anything..."
              className="flex-1 px-4 py-2 bg-white/5 rounded-xl border border-white/10 focus:border-blue-500/50 focus:outline-none transition-colors"
            />
            <button
              type="submit"
              disabled={!input.trim()}
              className="p-2 rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 hover:shadow-lg hover:shadow-purple-500/25 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Send className="w-5 h-5 text-white" />
            </button>
          </form>
        </div>
      </div>
    </>
  )
}
