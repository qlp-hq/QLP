'use client'

import React, { useState, useEffect } from 'react'
import { Sparkles, Loader2, Code2, Layers, GitBranch, Package, CheckCircle, ArrowRight, Wand2 } from 'lucide-react'

const examplePrompts = [
  "Build a real-time collaborative task management system with team workspaces and Slack integration",
  "Create an e-commerce platform with inventory management, payment processing, and shipping APIs",
  "Develop a social media analytics dashboard with real-time data streaming and ML insights",
  "Build a healthcare appointment booking system with patient records and video consultations",
  "Create a financial portfolio tracker with live market data and automated trading signals"
]

const generationSteps = [
  { icon: Wand2, label: "Analyzing Intent", color: "from-blue-500 to-cyan-500" },
  { icon: Layers, label: "Selecting AI Agents", color: "from-purple-500 to-pink-500" },
  { icon: Code2, label: "Generating Code", color: "from-green-500 to-emerald-500" },
  { icon: GitBranch, label: "Creating Architecture", color: "from-orange-500 to-red-500" },
  { icon: Package, label: "Building Services", color: "from-indigo-500 to-purple-500" },
  { icon: CheckCircle, label: "Validating Quality", color: "from-teal-500 to-green-500" }
]

export default function Playground() {
  const [input, setInput] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)
  const [currentStep, setCurrentStep] = useState(-1)
  const [generationComplete, setGenerationComplete] = useState(false)
  const [selectedPrompt, setSelectedPrompt] = useState<number | null>(null)

  const handleGenerate = () => {
    if (!input.trim()) return
    
    setIsGenerating(true)
    setGenerationComplete(false)
    setCurrentStep(0)

    // Simulate generation steps
    const stepDuration = 2000
    generationSteps.forEach((_, index) => {
      setTimeout(() => {
        setCurrentStep(index)
      }, index * stepDuration)
    })

    setTimeout(() => {
      setIsGenerating(false)
      setGenerationComplete(true)
      setCurrentStep(-1)
    }, generationSteps.length * stepDuration)
  }

  const selectExample = (index: number) => {
    setSelectedPrompt(index)
    setInput(examplePrompts[index])
    setGenerationComplete(false)
  }

  return (
    <section id="playground" className="py-32 relative overflow-hidden">
      {/* Background Elements */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-0 left-1/4 w-[500px] h-[500px] bg-blue-500 rounded-full blur-3xl" />
        <div className="absolute bottom-0 right-1/4 w-[500px] h-[500px] bg-purple-500 rounded-full blur-3xl" />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-5xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass border border-white/10 mb-6">
              <Sparkles className="w-4 h-4 text-purple-400" />
              <span className="text-sm font-medium">Interactive Demo</span>
            </div>
            
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              Experience the Magic of{' '}
              <span className="gradient-text">LLOM</span>
            </h2>
            
            <p className="text-xl text-gray-400 max-w-2xl mx-auto">
              Describe your software idea and watch as our AI orchestrates multiple specialized agents to build it in seconds
            </p>
          </div>

          {/* Main Playground */}
          <div className="glass rounded-3xl p-8 md:p-10 border border-white/10">
            {/* Input Section */}
            <div className="mb-8">
              <label className="block text-sm font-medium text-gray-300 mb-4">
                Describe your software in plain English
              </label>
              
              <div className="relative">
                <textarea
                  value={input}
                  onChange={(e) => {
                    setInput(e.target.value)
                    setSelectedPrompt(null)
                  }}
                  placeholder="e.g., Build a real-time collaborative task management system..."
                  className="w-full h-32 bg-black/50 rounded-2xl px-6 py-4 text-lg focus:outline-none focus:ring-2 focus:ring-blue-500/50 resize-none border border-white/10 focus:border-blue-500/50 transition-all duration-300 placeholder-gray-600"
                />
                
                {/* Character Count */}
                <div className="absolute bottom-4 right-4 text-xs text-gray-500">
                  {input.length} / 500
                </div>
              </div>
            </div>

            {/* Example Prompts */}
            <div className="mb-8">
              <p className="text-sm font-medium text-gray-400 mb-3">Try an example:</p>
              <div className="flex flex-wrap gap-2">
                {examplePrompts.map((prompt, index) => (
                  <button
                    key={index}
                    onClick={() => selectExample(index)}
                    className={`px-4 py-2 rounded-xl text-sm font-medium transition-all duration-300 ${
                      selectedPrompt === index
                        ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white'
                        : 'glass border border-white/10 hover:border-white/20 text-gray-300 hover:text-white'
                    }`}
                  >
                    Example {index + 1}
                  </button>
                ))}
              </div>
            </div>

            {/* Generation Button */}
            <button
              onClick={handleGenerate}
              disabled={isGenerating || !input.trim()}
              className="w-full group relative px-8 py-4 overflow-hidden rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold text-lg transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/25 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:shadow-none"
            >
              <span className="relative z-10 flex items-center justify-center gap-3">
                {isGenerating ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" />
                    Orchestrating AI Agents...
                  </>
                ) : (
                  <>
                    <Sparkles className="w-5 h-5" />
                    Generate Microservices
                    <ArrowRight className="w-5 h-5 transition-transform duration-300 group-hover:translate-x-1" />
                  </>
                )}
              </span>
              
              {/* Animated Background */}
              {!isGenerating && (
                <div className="absolute inset-0 bg-gradient-to-r from-blue-700 to-purple-700 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              )}
            </button>

            {/* Generation Progress */}
            {isGenerating && (
              <div className="mt-8 space-y-4">
                {generationSteps.map((step, index) => {
                  const Icon = step.icon
                  const isActive = currentStep === index
                  const isComplete = currentStep > index
                  
                  return (
                    <div
                      key={index}
                      className={`flex items-center gap-4 p-4 rounded-xl transition-all duration-500 ${
                        isActive
                          ? 'glass border border-white/20 scale-105'
                          : isComplete
                          ? 'opacity-60'
                          : 'opacity-30'
                      }`}
                    >
                      <div className={`p-3 rounded-xl bg-gradient-to-br ${step.color} ${
                        isActive ? 'animate-pulse' : ''
                      }`}>
                        <Icon className="w-5 h-5 text-white" />
                      </div>
                      
                      <div className="flex-1">
                        <p className={`font-medium ${isActive ? 'text-white' : 'text-gray-400'}`}>
                          {step.label}
                        </p>
                      </div>
                      
                      {isComplete && (
                        <CheckCircle className="w-5 h-5 text-green-400" />
                      )}
                      
                      {isActive && (
                        <div className="flex gap-1">
                          <div className="w-2 h-2 bg-white rounded-full animate-pulse" />
                          <div className="w-2 h-2 bg-white rounded-full animate-pulse delay-100" />
                          <div className="w-2 h-2 bg-white rounded-full animate-pulse delay-200" />
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            )}

            {/* Generation Complete */}
            {generationComplete && (
              <div className="mt-8 p-6 rounded-2xl bg-gradient-to-br from-green-500/10 to-emerald-500/10 border border-green-500/20">
                <div className="flex items-center gap-4 mb-4">
                  <div className="p-3 rounded-xl bg-gradient-to-br from-green-500 to-emerald-500">
                    <CheckCircle className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="text-xl font-semibold text-white">Generation Complete!</h3>
                    <p className="text-sm text-gray-400">Your microservices are ready in 12.5 seconds</p>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                  <div className="text-center">
                    <div className="text-2xl font-bold gradient-text mb-1">8</div>
                    <div className="text-xs text-gray-400">Microservices</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold gradient-text mb-1">2,847</div>
                    <div className="text-xs text-gray-400">Lines of Code</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold gradient-text mb-1">100%</div>
                    <div className="text-xs text-gray-400">Test Coverage</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold gradient-text mb-1">A+</div>
                    <div className="text-xs text-gray-400">Quality Score</div>
                  </div>
                </div>
                
                <button className="w-full px-6 py-3 rounded-xl bg-gradient-to-r from-green-600 to-emerald-600 text-white font-semibold hover:shadow-lg hover:shadow-green-500/25 transition-all duration-300">
                  View Generated Code
                </button>
              </div>
            )}
          </div>

          {/* Info Cards */}
          <div className="grid md:grid-cols-3 gap-6 mt-12">
            <div className="glass rounded-2xl p-6 border border-white/5 hover:border-white/10 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="p-2 rounded-lg bg-blue-500/20">
                  <Wand2 className="w-5 h-5 text-blue-400" />
                </div>
                <div>
                  <h3 className="font-semibold mb-2">Natural Language</h3>
                  <p className="text-sm text-gray-400">
                    No coding required. Just describe what you want to build in plain English.
                  </p>
                </div>
              </div>
            </div>
            
            <div className="glass rounded-2xl p-6 border border-white/5 hover:border-white/10 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="p-2 rounded-lg bg-purple-500/20">
                  <Layers className="w-5 h-5 text-purple-400" />
                </div>
                <div>
                  <h3 className="font-semibold mb-2">AI Orchestration</h3>
                  <p className="text-sm text-gray-400">
                    Multiple specialized AI agents work together to build your application.
                  </p>
                </div>
              </div>
            </div>
            
            <div className="glass rounded-2xl p-6 border border-white/5 hover:border-white/10 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="p-2 rounded-lg bg-green-500/20">
                  <Package className="w-5 h-5 text-green-400" />
                </div>
                <div>
                  <h3 className="font-semibold mb-2">Production Ready</h3>
                  <p className="text-sm text-gray-400">
                    Get complete microservices with tests, docs, and deployment configs.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
