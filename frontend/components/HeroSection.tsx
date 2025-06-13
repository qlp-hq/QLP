'use client'

import React, { useState, useEffect } from 'react'
import { ArrowRight, Sparkles, Zap, Shield, Globe, Play, ChevronDown } from 'lucide-react'

export default function HeroSection() {
  const [currentWord, setCurrentWord] = useState(0)
  const words = ['Revolutionary', 'Lightning-Fast', 'Enterprise-Ready', 'AI-Powered']
  
  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentWord((prev) => (prev + 1) % words.length)
    }, 3000)
    return () => clearInterval(interval)
  }, [])

  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden pt-20">
      {/* Animated Background Gradient */}
      <div className="absolute inset-0 opacity-30">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500 rounded-full blur-3xl animate-pulse" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-600 rounded-full blur-3xl animate-pulse delay-1000" />
        <div className="absolute top-1/2 left-1/2 w-96 h-96 bg-cyan-500 rounded-full blur-3xl animate-pulse delay-2000" />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-5xl mx-auto text-center">
          {/* Animated Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass border border-white/10 mb-8 group hover:border-white/20 transition-all duration-300">
            <div className="relative">
              <Sparkles className="w-4 h-4 text-blue-400" />
              <div className="absolute inset-0 animate-ping">
                <Sparkles className="w-4 h-4 text-blue-400 opacity-50" />
              </div>
            </div>
            <span className="text-sm font-medium text-gray-300">
              Introducing LLOM Architecture
            </span>
            <ArrowRight className="w-3 h-3 text-gray-400 group-hover:translate-x-0.5 transition-transform" />
          </div>

          {/* Main Headline */}
          <h1 className="text-5xl md:text-7xl lg:text-8xl font-bold mb-8 leading-tight">
            <span className="block text-white mb-2">Build</span>
            <span className="relative inline-block">
              <span className="gradient-text">{words[currentWord]}</span>
              <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg blur opacity-30 animate-pulse" />
            </span>
            <span className="block text-white mt-2">Microservices</span>
          </h1>

          {/* Timer Display */}
          <div className="flex items-center justify-center gap-4 mb-8">
            <div className="flex items-center gap-2 px-6 py-3 rounded-2xl glass border border-white/10">
              <Zap className="w-5 h-5 text-yellow-400" />
              <span className="text-3xl font-mono font-bold gradient-text">12.5</span>
              <span className="text-lg text-gray-400">seconds</span>
            </div>
          </div>

          {/* Subheadline */}
          <p className="text-xl md:text-2xl text-gray-400 max-w-3xl mx-auto mb-12 leading-relaxed">
            Transform your ideas into production-ready code with our{' '}
            <span className="text-white font-semibold">Large Language Orchestration Model</span>.
            Just describe what you want to build in plain English.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
            <button className="group relative px-8 py-4 overflow-hidden rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 text-white font-semibold text-lg transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/25 hover:scale-105">
              <span className="relative z-10 flex items-center gap-2">
                Start Building Now
                <ArrowRight className="w-5 h-5 transition-transform duration-300 group-hover:translate-x-1" />
              </span>
              
              {/* Animated Background */}
              <div className="absolute inset-0 bg-gradient-to-r from-blue-700 to-purple-700 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
              
              {/* Shimmer Effect */}
              <div className="absolute inset-0 -top-[40px] bg-gradient-to-r from-transparent via-white/10 to-transparent skew-x-12 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-1000" />
            </button>

            <button className="group px-8 py-4 rounded-2xl glass border border-white/10 hover:border-white/20 font-semibold text-lg transition-all duration-300 hover:bg-white/5">
              <span className="flex items-center gap-2">
                <Play className="w-5 h-5" />
                Watch Demo
                <span className="text-sm text-gray-500">(2 min)</span>
              </span>
            </button>
          </div>

          {/* Trust Indicators */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
            <div className="glass rounded-2xl p-6 border border-white/5 hover:border-white/10 transition-all duration-300 group">
              <div className="flex items-center justify-center mb-4">
                <div className="p-3 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 group-hover:from-blue-500/30 group-hover:to-purple-500/30 transition-all duration-300">
                  <Globe className="w-6 h-6 text-blue-400" />
                </div>
              </div>
              <div className="text-3xl font-bold mb-1 gradient-text">10M+</div>
              <div className="text-sm text-gray-400">Code Patterns Indexed</div>
            </div>

            <div className="glass rounded-2xl p-6 border border-white/5 hover:border-white/10 transition-all duration-300 group">
              <div className="flex items-center justify-center mb-4">
                <div className="p-3 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 group-hover:from-green-500/30 group-hover:to-emerald-500/30 transition-all duration-300">
                  <Shield className="w-6 h-6 text-green-400" />
                </div>
              </div>
              <div className="text-3xl font-bold mb-1 gradient-text">94%</div>
              <div className="text-sm text-gray-400">Quality Score Average</div>
            </div>

            <div className="glass rounded-2xl p-6 border border-white/5 hover:border-white/10 transition-all duration-300 group">
              <div className="flex items-center justify-center mb-4">
                <div className="p-3 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 group-hover:from-purple-500/30 group-hover:to-pink-500/30 transition-all duration-300">
                  <Sparkles className="w-6 h-6 text-purple-400" />
                </div>
              </div>
              <div className="text-3xl font-bold mb-1 gradient-text">16+</div>
              <div className="text-sm text-gray-400">Specialized AI Agents</div>
            </div>
          </div>
        </div>
      </div>

      {/* Scroll Indicator */}
      <div className="absolute bottom-8 left-1/2 transform -translate-x-1/2 animate-bounce">
        <ChevronDown className="w-6 h-6 text-gray-400" />
      </div>
    </section>
  )
}
