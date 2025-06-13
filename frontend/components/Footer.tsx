'use client'

import React from 'react'
import { Github, Twitter, Linkedin, Youtube, Mail, MapPin, Sparkles } from 'lucide-react'
import Link from 'next/link'

export default function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="relative pt-20 pb-10 border-t border-white/5">
      {/* Background Gradient */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-blue-500 rounded-full blur-3xl" />
        <div className="absolute bottom-0 right-0 w-96 h-96 bg-purple-600 rounded-full blur-3xl" />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="grid md:grid-cols-4 gap-8 mb-12">
          {/* Company Info */}
          <div className="md:col-span-1">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 relative">
                <div className="absolute inset-0 rounded-xl bg-gradient-to-br from-blue-500 to-purple-600" />
                <div className="absolute inset-1 rounded-lg bg-black flex items-center justify-center">
                  <span className="text-xl font-bold bg-gradient-to-br from-blue-400 to-purple-500 bg-clip-text text-transparent">
                    Q
                  </span>
                </div>
              </div>
              <div>
                <span className="text-xl font-bold">QuantumLayer</span>
              </div>
            </div>
            
            <p className="text-sm text-gray-400 mb-4">
              Revolutionizing software development with AI-powered microservice generation.
            </p>
            
            <div className="flex items-center gap-2 text-sm text-gray-500 mb-2">
              <MapPin className="w-4 h-4" />
              <span>London, UK</span>
            </div>
            
            <div className="flex items-center gap-2 text-sm text-gray-500">
              <Mail className="w-4 h-4" />
              <span>hello@quantumlayer.ai</span>
            </div>
          </div>

          {/* Product */}
          <div>
            <h3 className="font-semibold mb-4">Product</h3>
            <ul className="space-y-2">
              <li>
                <Link href="#features" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Features
                </Link>
              </li>
              <li>
                <Link href="#pricing" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Pricing
                </Link>
              </li>
              <li>
                <Link href="#playground" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Playground
                </Link>
              </li>
              <li>
                <Link href="#api" className="text-sm text-gray-400 hover:text-white transition-colors">
                  API Access
                </Link>
              </li>
              <li>
                <Link href="#enterprise" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Enterprise
                </Link>
              </li>
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h3 className="font-semibold mb-4">Resources</h3>
            <ul className="space-y-2">
              <li>
                <Link href="#docs" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Documentation
                </Link>
              </li>
              <li>
                <Link href="#tutorials" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Tutorials
                </Link>
              </li>
              <li>
                <Link href="#blog" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Blog
                </Link>
              </li>
              <li>
                <Link href="#community" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Community
                </Link>
              </li>
              <li>
                <Link href="#changelog" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Changelog
                </Link>
              </li>
            </ul>
          </div>

          {/* Company */}
          <div>
            <h3 className="font-semibold mb-4">Company</h3>
            <ul className="space-y-2">
              <li>
                <Link href="#about" className="text-sm text-gray-400 hover:text-white transition-colors">
                  About Us
                </Link>
              </li>
              <li>
                <Link href="#careers" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Careers
                  <span className="ml-2 px-2 py-0.5 text-xs bg-green-500/20 text-green-400 rounded-full">
                    We're hiring!
                  </span>
                </Link>
              </li>
              <li>
                <Link href="#press" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Press Kit
                </Link>
              </li>
              <li>
                <Link href="#privacy" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Privacy Policy
                </Link>
              </li>
              <li>
                <Link href="#terms" className="text-sm text-gray-400 hover:text-white transition-colors">
                  Terms of Service
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Newsletter */}
        <div className="glass rounded-2xl p-6 mb-12 border border-white/10">
          <div className="md:flex items-center justify-between">
            <div className="mb-4 md:mb-0">
              <h3 className="font-semibold mb-2 flex items-center gap-2">
                <Sparkles className="w-5 h-5 text-purple-400" />
                Stay in the Quantum Loop
              </h3>
              <p className="text-sm text-gray-400">
                Get updates on new features and AI advancements
              </p>
            </div>
            
            <form className="flex gap-2">
              <input
                type="email"
                placeholder="Enter your email"
                className="px-4 py-2 bg-black/50 rounded-xl border border-white/10 focus:border-blue-500/50 focus:outline-none transition-colors text-sm"
              />
              <button
                type="submit"
                className="px-6 py-2 bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl font-medium text-sm hover:shadow-lg hover:shadow-purple-500/25 transition-all duration-300"
              >
                Subscribe
              </button>
            </form>
          </div>
        </div>

        {/* Bottom Section */}
        <div className="flex flex-col md:flex-row items-center justify-between pt-8 border-t border-white/5">
          <div className="text-sm text-gray-400 mb-4 md:mb-0">
            © {currentYear} QuantumLayer. All rights reserved.
          </div>
          
          {/* Social Links */}
          <div className="flex items-center gap-4">
            <a
              href="https://github.com"
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 rounded-lg hover:bg-white/10 transition-colors"
            >
              <Github className="w-5 h-5 text-gray-400 hover:text-white" />
            </a>
            <a
              href="https://twitter.com"
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 rounded-lg hover:bg-white/10 transition-colors"
            >
              <Twitter className="w-5 h-5 text-gray-400 hover:text-white" />
            </a>
            <a
              href="https://linkedin.com"
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 rounded-lg hover:bg-white/10 transition-colors"
            >
              <Linkedin className="w-5 h-5 text-gray-400 hover:text-white" />
            </a>
            <a
              href="https://youtube.com"
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 rounded-lg hover:bg-white/10 transition-colors"
            >
              <Youtube className="w-5 h-5 text-gray-400 hover:text-white" />
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
