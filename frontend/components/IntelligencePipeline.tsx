'use client'

import React, { useState, useEffect } from 'react'
import { Brain, Target, Users, Code2, Shield, Package, ArrowRight, Sparkles } from 'lucide-react'

const pipelineSteps = [
  {
    icon: Target,
    title: "Intent Analysis",
    description: "AI understands your requirements and project goals",
    color: "from-blue-500 to-cyan-500",
    details: [
      "Natural language processing",
      "Context understanding",
      "Requirement extraction",
      "Technology recommendation"
    ]
  },
  {
    icon: Users,
    title: "Agent Selection",
    description: "LLOM selects the perfect team of AI specialists",
    color: "from-purple-500 to-pink-500",
    details: [
      "Dynamic agent orchestration",
      "Skill-based selection",
      "6-16 specialized agents",
      "Real-time coordination"
    ]
  },
  {
    icon: Brain,
    title: "Intelligent Orchestration",
    description: "Multiple AI agents collaborate on your project",
    color: "from-indigo-500 to-purple-500",
    details: [
      "Parallel processing",
      "Inter-agent communication",
      "Decision optimization",
      "Conflict resolution"
    ]
  },
  {
    icon: Code2,
    title: "Code Generation",
    description: "Production-ready code is generated from scratch",
    color: "from-green-500 to-emerald-500",
    details: [
      "Clean architecture",
      "Best practices",
      "Modern frameworks",
      "Optimized performance"
    ]
  },
  {
    icon: Shield,
    title: "Quality Validation",
    description: "Automated testing and security scanning",
    color: "from-orange-500 to-red-500",
    details: [
      "Unit & integration tests",
      "Security analysis",
      "Performance benchmarks",
      "Code quality metrics"
    ]
  },
  {
    icon: Package,
    title: "Deployment Ready",
    description: "Complete package with everything you need",
    color: "from-teal-500 to-cyan-500",
    details: [
      "Docker containers",
      "CI/CD pipelines",
      "Documentation",
      "Monitoring setup"
    ]
  }
]

export default function IntelligencePipeline() {
  const [activeStep, setActiveStep] = useState(0)
  const [isAnimating, setIsAnimating] = useState(true)

  useEffect(() => {
    if (!isAnimating) return

    const interval = setInterval(() => {
      setActiveStep((prev) => (prev + 1) % pipelineSteps.length)
    }, 3000)

    return () => clearInterval(interval)
  }, [isAnimating])

  return (
    <section id="how-it-works" className="py-32 relative overflow-hidden">
      {/* Animated Background */}
      <div className="absolute inset-0 opacity-20">
        <div className="absolute top-1/2 left-0 w-full h-1 bg-gradient-to-r from-transparent via-blue-500 to-transparent animate-pulse" />
        <div className="absolute top-1/2 left-0 w-full h-px bg-gradient-to-r from-transparent via-purple-500 to-transparent animate-pulse delay-1000" />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-6xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass border border-white/10 mb-6">
              <Brain className="w-4 h-4 text-indigo-400" />
              <span className="text-sm font-medium">Intelligence Pipeline</span>
            </div>
            
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              How QuantumLayer{' '}
              <span className="gradient-text">Orchestrates AI</span>
            </h2>
            
            <p className="text-xl text-gray-400 max-w-3xl mx-auto">
              Our 6-stage intelligence pipeline ensures every project is built with precision,
              quality, and blazing speed
            </p>
          </div>

          {/* Pipeline Visualization */}
          <div className="relative mb-16">
            {/* Connection Line */}
            <div className="absolute top-1/2 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-500/20 via-purple-500/20 to-cyan-500/20 -translate-y-1/2 hidden md:block" />
            
            {/* Steps */}
            <div className="grid md:grid-cols-6 gap-4 relative">
              {pipelineSteps.map((step, index) => {
                const Icon = step.icon
                const isActive = activeStep === index
                const isPast = activeStep > index
                
                return (
                  <div
                    key={index}
                    className="relative"
                    onMouseEnter={() => {
                      setIsAnimating(false)
                      setActiveStep(index)
                    }}
                    onMouseLeave={() => setIsAnimating(true)}
                  >
                    <div className={`glass rounded-xl p-4 border transition-all duration-300 cursor-pointer ${
                      isActive
                        ? 'border-white/20 scale-105 shadow-xl'
                        : 'border-white/5 hover:border-white/10'
                    }`}>
                      {/* Step Number */}
                      <div className="absolute -top-3 -right-3 w-8 h-8 rounded-full bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center text-sm font-bold">
                        {index + 1}
                      </div>
                      
                      {/* Icon */}
                      <div className={`p-3 rounded-lg bg-gradient-to-br ${step.color} mb-3 ${
                        isActive ? 'animate-pulse' : ''
                      }`}>
                        <Icon className="w-6 h-6 text-white" />
                      </div>
                      
                      {/* Content */}
                      <h3 className="font-semibold text-sm mb-1">{step.title}</h3>
                      <p className="text-xs text-gray-400 line-clamp-2">{step.description}</p>
                      
                      {/* Progress Indicator */}
                      {(isActive || isPast) && (
                        <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-500 to-purple-500 rounded-b-xl" />
                      )}
                    </div>
                    
                    {/* Arrow */}
                    {index < pipelineSteps.length - 1 && (
                      <div className="hidden md:block absolute top-1/2 -right-2 -translate-y-1/2">
                        <ArrowRight className={`w-4 h-4 transition-colors duration-300 ${
                          isPast || isActive ? 'text-blue-400' : 'text-gray-600'
                        }`} />
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>

          {/* Active Step Details */}
          <div className="glass rounded-3xl p-8 md:p-10 border border-white/10">
            <div className="grid md:grid-cols-2 gap-8 items-center">
              <div>
                <div className="flex items-center gap-4 mb-6">
                  <div className={`p-4 rounded-xl bg-gradient-to-br ${pipelineSteps[activeStep].color}`}>
                    {React.createElement(pipelineSteps[activeStep].icon, {
                      className: "w-8 h-8 text-white"
                    })}
                  </div>
                  <div>
                    <h3 className="text-2xl font-bold">
                      Step {activeStep + 1}: {pipelineSteps[activeStep].title}
                    </h3>
                    <p className="text-gray-400">{pipelineSteps[activeStep].description}</p>
                  </div>
                </div>
                
                <div className="space-y-3">
                  {pipelineSteps[activeStep].details.map((detail, index) => (
                    <div key={index} className="flex items-center gap-3">
                      <div className="w-2 h-2 rounded-full bg-gradient-to-r from-blue-400 to-purple-400" />
                      <span className="text-sm text-gray-300">{detail}</span>
                    </div>
                  ))}
                </div>
              </div>
              
              {/* Visual Representation */}
              <div className="relative h-64 flex items-center justify-center">
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className={`w-48 h-48 rounded-full bg-gradient-to-br ${pipelineSteps[activeStep].color} opacity-20 animate-pulse`} />
                </div>
                <div className="relative">
                  <div className={`p-8 rounded-2xl bg-gradient-to-br ${pipelineSteps[activeStep].color}`}>
                    {React.createElement(pipelineSteps[activeStep].icon, {
                      className: "w-16 h-16 text-white"
                    })}
                  </div>
                  <div className="absolute -inset-8">
                    <div className={`w-32 h-32 border-4 border-white/20 rounded-full animate-spin-slow`} />
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Stats */}
          <div className="grid md:grid-cols-4 gap-6 mt-12">
            <div className="text-center">
              <div className="text-3xl font-bold gradient-text mb-2">12.5s</div>
              <div className="text-sm text-gray-400">Average Pipeline Time</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold gradient-text mb-2">6-16</div>
              <div className="text-sm text-gray-400">AI Agents per Project</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold gradient-text mb-2">94%</div>
              <div className="text-sm text-gray-400">Quality Score</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold gradient-text mb-2">100%</div>
              <div className="text-sm text-gray-400">Test Coverage</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
