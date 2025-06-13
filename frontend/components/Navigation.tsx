'use client'

import React, { useState, useEffect } from 'react'
import { Menu, X, ChevronRight, Sparkles } from 'lucide-react'
import Link from 'next/link'

export default function Navigation() {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20)
    }
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  return (
    <nav className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
      scrolled ? 'py-4' : 'py-6'
    }`}>
      <div className={`mx-6 rounded-2xl transition-all duration-300 ${
        scrolled ? 'glass-darker shadow-2xl shadow-black/50' : ''
      }`}>
        <div className="container mx-auto px-6">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <Link href="/" className="flex items-center gap-3 group">
              <div className="relative">
                {/* Animated Logo */}
                <div className="w-10 h-10 relative">
                  {/* Outer Ring */}
                  <div className="absolute inset-0 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600 animate-pulse" />
                  
                  {/* Inner Core */}
                  <div className="absolute inset-1 rounded-lg bg-black flex items-center justify-center">
                    <div className="relative">
                      {/* Q Letter */}
                      <span className="text-xl font-bold bg-gradient-to-br from-blue-400 to-purple-500 bg-clip-text text-transparent">
                        Q
                      </span>
                      
                      {/* Orbiting Particle */}
                      <div className="absolute -top-1 -right-1 w-2 h-2 bg-cyan-400 rounded-full animate-pulse" />
                    </div>
                  </div>
                  
                  {/* Glow Effect */}
                  <div className="absolute inset-0 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-600/20 blur-xl group-hover:blur-2xl transition-all duration-300" />
                </div>
              </div>
              
              <div className="flex flex-col">
                <span className="text-xl font-bold tracking-tight">
                  Quantum<span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-purple-500">Layer</span>
                </span>
                <span className="text-[10px] text-gray-500 tracking-widest uppercase">
                  Next-Gen AI Platform
                </span>
              </div>
            </Link>

            {/* Desktop Menu */}
            <div className="hidden md:flex items-center gap-8">
              <Link href="#how-it-works" className="text-gray-300 hover:text-white transition-colors duration-200 text-sm font-medium">
                How it Works
              </Link>
              <Link href="#features" className="text-gray-300 hover:text-white transition-colors duration-200 text-sm font-medium">
                Features
              </Link>
              <Link href="#pricing" className="text-gray-300 hover:text-white transition-colors duration-200 text-sm font-medium">
                Pricing
              </Link>
              <Link href="#docs" className="text-gray-300 hover:text-white transition-colors duration-200 text-sm font-medium">
                Documentation
              </Link>
              <Link href="#blog" className="text-gray-300 hover:text-white transition-colors duration-200 text-sm font-medium">
                Blog
              </Link>
            </div>

            {/* CTA Buttons */}
            <div className="hidden md:flex items-center gap-4">
              <button className="text-gray-300 hover:text-white transition-colors duration-200 text-sm font-medium">
                Sign In
              </button>
              
              <button className="group relative px-5 py-2.5 overflow-hidden rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 text-white font-medium text-sm transition-all duration-300 hover:shadow-lg hover:shadow-purple-500/25">
                <span className="relative z-10 flex items-center gap-2">
                  Start Building
                  <ChevronRight className="w-4 h-4 transition-transform duration-300 group-hover:translate-x-0.5" />
                </span>
                
                {/* Shimmer Effect */}
                <div className="absolute inset-0 -top-[20px] bg-gradient-to-r from-transparent via-white/20 to-transparent skew-x-12 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-1000" />
              </button>
            </div>

            {/* Mobile Menu Button */}
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="md:hidden p-2 rounded-lg hover:bg-white/10 transition-colors duration-200"
            >
              {isMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      <div className={`md:hidden fixed inset-x-0 top-20 mx-6 rounded-2xl glass-darker shadow-2xl shadow-black/50 transition-all duration-300 transform ${
        isMenuOpen ? 'translate-y-0 opacity-100' : '-translate-y-4 opacity-0 pointer-events-none'
      }`}>
        <div className="p-6 space-y-4">
          <Link href="#how-it-works" className="block text-gray-300 hover:text-white transition-colors duration-200 font-medium">
            How it Works
          </Link>
          <Link href="#features" className="block text-gray-300 hover:text-white transition-colors duration-200 font-medium">
            Features
          </Link>
          <Link href="#pricing" className="block text-gray-300 hover:text-white transition-colors duration-200 font-medium">
            Pricing
          </Link>
          <Link href="#docs" className="block text-gray-300 hover:text-white transition-colors duration-200 font-medium">
            Documentation
          </Link>
          <Link href="#blog" className="block text-gray-300 hover:text-white transition-colors duration-200 font-medium">
            Blog
          </Link>
          
          <div className="pt-4 space-y-3 border-t border-white/10">
            <button className="w-full text-center text-gray-300 hover:text-white transition-colors duration-200 font-medium">
              Sign In
            </button>
            <button className="w-full px-5 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 text-white font-medium transition-all duration-300">
              Start Building
            </button>
          </div>
        </div>
      </div>
    </nav>
  )
}
