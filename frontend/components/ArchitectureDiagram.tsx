'use client'

import React, { useState } from 'react'
import { Brain, Code2, Database, Shield, Zap, GitBranch, Cloud, Globe, Cpu, Layers, ArrowRight, CheckCircle } from 'lucide-react'

const agents = [
  { icon: Code2, name: "Frontend Architect", color: "from-blue-500 to-cyan-500", description: "React, Vue, Angular expertise" },
  { icon: Database, name: "Database Architect", color: "from-green-500 to-emerald-500", description: "SQL, NoSQL, Vector DBs" },
  { icon: Shield, name: "Security Expert", color: "from-red-500 to-orange-500", description: "Auth, Encryption, Best Practices" },
  { icon: Cloud, name: "DevOps Specialist", color: "from-purple-500 to-pink-500", description: "CI/CD, Kubernetes, Docker" },
  { icon: Zap, name: "Performance Guru", color: "from-yellow-500 to-amber-500", description: "Optimization, Caching, Scaling" },
  { icon: GitBranch, name: "Backend Engineer", color: "from-indigo-500 to-purple-500", description: "APIs, Microservices, Events" }
]

export default function ArchitectureDiagram() {
  const [selectedAgent, setSelectedAgent] = useState<number | null>(null)
  const [hoveredSection, setHoveredSection] = useState<string | null>(null)

  return (
    <section className="py-32 relative overflow-hidden">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `radial-gradient(circle at 2px 2px, white 1px, transparent 1px)`,
          backgroundSize: '40px 40px'
        }} />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-6xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass border border-white/10 mb-6">
              <Brain className="w-4 h-4 text-blue-400" />
              <span className="text-sm font-medium">LLOM Architecture</span>
            </div>
            
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              Large Language{' '}
              <span className="gradient-text">Orchestration Model</span>
            </h2>
            
            <p className="text-xl text-gray-400 max-w-3xl mx-auto">
              Unlike traditional code generators, LLOM dynamically orchestrates specialized AI agents
              to create unique, production-ready solutions tailored to your specific needs
            </p>
          </div>

          {/* Main Architecture Visualization */}
          <div className="glass rounded-3xl p-8 md:p-12 border border-white/10">
            <div className="grid md:grid-cols-3 gap-8">
              {/* Input Section */}
              <div 
                className="relative"
                onMouseEnter={() => setHoveredSection('input')}
                onMouseLeave={() => setHoveredSection(null)}
              >
                <div className={`glass rounded-2xl p-6 border transition-all duration-300 ${
                  hoveredSection === 'input' ? 'border-blue-500/50 scale-105' : 'border-white/10'
                }`}>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="p-3 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20">
                      <Globe className="w-6 h-6 text-blue-400" />
                    </div>
                    <h3 className="text-lg font-semibold">Natural Language Input</h3>
                  </div>
                  
                  <p className="text-sm text-gray-400 mb-4">
                    Describe your software requirements in plain English
                  </p>
                  
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Intent Analysis</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Context Understanding</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Requirement Extraction</span>
                    </div>
                  </div>
                </div>
                
                {/* Arrow */}
                <div className="hidden md:block absolute top-1/2 -right-4 transform -translate-y-1/2">
                  <ArrowRight className="w-8 h-8 text-gray-600" />
                </div>
              </div>

              {/* LLOM Core */}
              <div 
                className="relative"
                onMouseEnter={() => setHoveredSection('llom')}
                onMouseLeave={() => setHoveredSection(null)}
              >
                <div className={`glass rounded-2xl p-6 border transition-all duration-300 ${
                  hoveredSection === 'llom' ? 'border-purple-500/50 scale-105' : 'border-white/10'
                }`}>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="p-3 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20">
                      <Brain className="w-6 h-6 text-purple-400" />
                    </div>
                    <h3 className="text-lg font-semibold">LLOM Orchestration</h3>
                  </div>
                  
                  <p className="text-sm text-gray-400 mb-4">
                    Intelligent agent selection and coordination
                  </p>
                  
                  {/* Central Brain Visual */}
                  <div className="relative h-40 flex items-center justify-center">
                    <div className="absolute inset-0 flex items-center justify-center">
                      <div className="w-24 h-24 rounded-full bg-gradient-to-br from-purple-600 to-pink-600 opacity-20 animate-pulse" />
                    </div>
                    <div className="relative">
                      <Cpu className="w-12 h-12 text-purple-400" />
                      <div className="absolute -inset-4">
                        <div className="w-20 h-20 border-2 border-purple-500/30 rounded-full animate-spin-slow" />
                      </div>
                    </div>
                  </div>
                  
                  <div className="text-center mt-4">
                    <span className="text-xs text-gray-500">Dynamic Agent Selection</span>
                  </div>
                </div>
                
                {/* Arrow */}
                <div className="hidden md:block absolute top-1/2 -right-4 transform -translate-y-1/2">
                  <ArrowRight className="w-8 h-8 text-gray-600" />
                </div>
              </div>

              {/* Output Section */}
              <div 
                className="relative"
                onMouseEnter={() => setHoveredSection('output')}
                onMouseLeave={() => setHoveredSection(null)}
              >
                <div className={`glass rounded-2xl p-6 border transition-all duration-300 ${
                  hoveredSection === 'output' ? 'border-green-500/50 scale-105' : 'border-white/10'
                }`}>
                  <div className="flex items-center gap-3 mb-4">
                    <div className="p-3 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20">
                      <Layers className="w-6 h-6 text-green-400" />
                    </div>
                    <h3 className="text-lg font-semibold">Production Output</h3>
                  </div>
                  
                  <p className="text-sm text-gray-400 mb-4">
                    Complete microservices architecture
                  </p>
                  
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Source Code</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Test Suites</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Documentation</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-sm">Deployment Configs</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* AI Agents Grid */}
          <div className="mt-16">
            <h3 className="text-2xl font-bold text-center mb-8">
              Specialized AI Agents
            </h3>
            
            <div className="grid md:grid-cols-3 gap-6">
              {agents.map((agent, index) => {
                const Icon = agent.icon
                const isSelected = selectedAgent === index
                
                return (
                  <button
                    key={index}
                    onClick={() => setSelectedAgent(isSelected ? null : index)}
                    className={`glass rounded-2xl p-6 border transition-all duration-300 text-left ${
                      isSelected
                        ? 'border-white/20 scale-105'
                        : 'border-white/5 hover:border-white/10'
                    }`}
                  >
                    <div className="flex items-start gap-4">
                      <div className={`p-3 rounded-xl bg-gradient-to-br ${agent.color} ${
                        isSelected ? 'scale-110' : ''
                      } transition-transform duration-300`}>
                        <Icon className="w-6 h-6 text-white" />
                      </div>
                      
                      <div className="flex-1">
                        <h4 className="font-semibold mb-1">{agent.name}</h4>
                        <p className="text-sm text-gray-400">{agent.description}</p>
                        
                        {isSelected && (
                          <div className="mt-3 text-xs text-gray-500">
                            This agent specializes in creating robust, scalable solutions
                            using industry best practices and modern technologies.
                          </div>
                        )}
                      </div>
                    </div>
                  </button>
                )
              })}
            </div>
          </div>

          {/* Key Differentiators */}
          <div className="mt-16 glass rounded-2xl p-8 border border-white/10">
            <h3 className="text-xl font-semibold mb-6 text-center">
              Why LLOM is Revolutionary
            </h3>
            
            <div className="grid md:grid-cols-3 gap-6">
              <div className="text-center">
                <div className="inline-flex p-3 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 mb-4">
                  <Zap className="w-6 h-6 text-blue-400" />
                </div>
                <h4 className="font-semibold mb-2">Not Template-Based</h4>
                <p className="text-sm text-gray-400">
                  Every output is uniquely generated from scratch based on your specific requirements
                </p>
              </div>
              
              <div className="text-center">
                <div className="inline-flex p-3 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 mb-4">
                  <Brain className="w-6 h-6 text-purple-400" />
                </div>
                <h4 className="font-semibold mb-2">Intelligent Coordination</h4>
                <p className="text-sm text-gray-400">
                  Multiple AI agents work together, each contributing their specialized expertise
                </p>
              </div>
              
              <div className="text-center">
                <div className="inline-flex p-3 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 mb-4">
                  <Shield className="w-6 h-6 text-green-400" />
                </div>
                <h4 className="font-semibold mb-2">Enterprise Quality</h4>
                <p className="text-sm text-gray-400">
                  Human architects validate critical decisions ensuring production-grade output
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
