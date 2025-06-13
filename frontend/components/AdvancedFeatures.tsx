'use client'

import React, { useState } from 'react'
import { 
  Rocket, Globe, Shield, Cpu, Database, GitBranch, 
  Zap, Lock, Cloud, BarChart, Users, Sparkles,
  ArrowRight, CheckCircle, TrendingUp, Server
} from 'lucide-react'

const features = [
  {
    icon: Rocket,
    title: "Instant Deployment",
    description: "Deploy to any cloud provider with one click. Pre-configured for AWS, Azure, GCP, and more.",
    category: "deployment",
    highlight: true,
  },
  {
    icon: Shield,
    title: "Enterprise Security",
    description: "Built-in authentication, encryption, and security best practices from day one.",
    category: "security",
  },
  {
    icon: Database,
    title: "Smart Data Layer",
    description: "Automatic database schema design with migrations, indexing, and optimization.",
    category: "data",
  },
  {
    icon: GitBranch,
    title: "Version Control",
    description: "Git-ready with proper branching strategies and CI/CD pipelines configured.",
    category: "development",
  },
  {
    icon: Zap,
    title: "Performance First",
    description: "Optimized code with caching, lazy loading, and performance monitoring built-in.",
    category: "performance",
  },
  {
    icon: Globe,
    title: "Global Scale",
    description: "Multi-region support with CDN integration and edge computing capabilities.",
    category: "scale",
  },
  {
    icon: Lock,
    title: "Compliance Ready",
    description: "GDPR, HIPAA, and SOC2 compliant architectures available out of the box.",
    category: "security",
  },
  {
    icon: Cloud,
    title: "Cloud Native",
    description: "Kubernetes-ready microservices with service mesh and observability.",
    category: "deployment",
  },
  {
    icon: BarChart,
    title: "Analytics Built-in",
    description: "Comprehensive monitoring, logging, and analytics from the start.",
    category: "monitoring",
  },
  {
    icon: Users,
    title: "Team Collaboration",
    description: "Multi-developer support with proper access controls and documentation.",
    category: "development",
  },
  {
    icon: Cpu,
    title: "AI Integration",
    description: "Easy integration with OpenAI, Anthropic, and other AI services.",
    category: "ai",
    highlight: true,
  },
  {
    icon: Server,
    title: "Infrastructure as Code",
    description: "Terraform and CloudFormation templates for complete infrastructure automation.",
    category: "deployment",
  },
]

const categories = [
  { id: 'all', label: 'All Features', icon: Sparkles },
  { id: 'deployment', label: 'Deployment', icon: Rocket },
  { id: 'security', label: 'Security', icon: Shield },
  { id: 'performance', label: 'Performance', icon: Zap },
  { id: 'development', label: 'Development', icon: Code2 },
]

export default function AdvancedFeatures() {
  const [selectedCategory, setSelectedCategory] = useState('all')
  const [hoveredFeature, setHoveredFeature] = useState<number | null>(null)

  const filteredFeatures = selectedCategory === 'all' 
    ? features 
    : features.filter(f => f.category === selectedCategory)

  return (
    <section id="features" className="py-32 relative overflow-hidden">
      {/* Background Elements */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-1/4 right-0 w-96 h-96 bg-purple-500 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 left-0 w-96 h-96 bg-blue-500 rounded-full blur-3xl" />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-6xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass border border-white/10 mb-6">
              <Sparkles className="w-4 h-4 text-purple-400" />
              <span className="text-sm font-medium">Enterprise Features</span>
            </div>
            
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              Everything You Need,{' '}
              <span className="gradient-text">Nothing You Don't</span>
            </h2>
            
            <p className="text-xl text-gray-400 max-w-3xl mx-auto">
              QuantumLayer generates complete, production-ready applications with 
              enterprise features built-in from the start
            </p>
          </div>

          {/* Category Filter */}
          <div className="flex flex-wrap justify-center gap-2 mb-12">
            {categories.map((category) => {
              const Icon = category.icon
              return (
                <button
                  key={category.id}
                  onClick={() => setSelectedCategory(category.id)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-xl transition-all duration-300 ${
                    selectedCategory === category.id
                      ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg shadow-purple-500/25'
                      : 'glass border border-white/10 hover:border-white/20 text-gray-300'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  <span className="text-sm font-medium">{category.label}</span>
                </button>
              )
            })}
          </div>

          {/* Features Grid */}
          <div className="grid md:grid-cols-3 gap-6 mb-16">
            {filteredFeatures.map((feature, index) => {
              const Icon = feature.icon
              const isHovered = hoveredFeature === index
              
              return (
                <div
                  key={index}
                  onMouseEnter={() => setHoveredFeature(index)}
                  onMouseLeave={() => setHoveredFeature(null)}
                  className={`glass rounded-2xl p-6 border transition-all duration-300 cursor-pointer ${
                    feature.highlight
                      ? 'border-purple-500/30 bg-gradient-to-br from-purple-500/5 to-blue-500/5'
                      : 'border-white/5 hover:border-white/10'
                  } ${isHovered ? 'scale-105 shadow-xl' : ''}`}
                >
                  <div className="flex items-start gap-4">
                    <div className={`p-3 rounded-xl transition-all duration-300 ${
                      isHovered
                        ? 'bg-gradient-to-br from-blue-500 to-purple-600 scale-110'
                        : 'bg-white/10'
                    }`}>
                      <Icon className="w-6 h-6 text-white" />
                    </div>
                    
                    <div className="flex-1">
                      <h3 className="font-semibold mb-2 flex items-center gap-2">
                        {feature.title}
                        {feature.highlight && (
                          <span className="px-2 py-0.5 text-xs bg-purple-500/20 text-purple-400 rounded-full">
                            Popular
                          </span>
                        )}
                      </h3>
                      <p className="text-sm text-gray-400 leading-relaxed">
                        {feature.description}
                      </p>
                    </div>
                  </div>
                  
                  {isHovered && (
                    <div className="mt-4 pt-4 border-t border-white/10">
                      <button className="text-sm text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-1">
                        Learn more
                        <ArrowRight className="w-3 h-3" />
                      </button>
                    </div>
                  )}
                </div>
              )
            })}
          </div>

          {/* Enterprise CTA */}
          <div className="glass rounded-3xl p-8 md:p-10 border border-white/10 bg-gradient-to-br from-blue-500/5 to-purple-500/5">
            <div className="grid md:grid-cols-2 gap-8 items-center">
              <div>
                <div className="flex items-center gap-2 mb-4">
                  <TrendingUp className="w-6 h-6 text-green-400" />
                  <span className="text-sm font-medium text-green-400">Enterprise Ready</span>
                </div>
                
                <h3 className="text-2xl font-bold mb-4">
                  Scale Without Limits
                </h3>
                
                <p className="text-gray-400 mb-6">
                  QuantumLayer Enterprise includes dedicated support, custom AI agents, 
                  on-premise deployment options, and SLAs that match your business needs.
                </p>
                
                <div className="space-y-2 mb-6">
                  <div className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-400" />
                    <span className="text-sm">Unlimited API calls</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-400" />
                    <span className="text-sm">Priority AI agent access</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-400" />
                    <span className="text-sm">24/7 dedicated support</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-green-400" />
                    <span className="text-sm">Custom agent training</span>
                  </div>
                </div>
                
                <button className="group px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl font-semibold hover:shadow-lg hover:shadow-purple-500/25 transition-all duration-300">
                  <span className="flex items-center gap-2">
                    Contact Sales
                    <ArrowRight className="w-4 h-4 transition-transform duration-300 group-hover:translate-x-1" />
                  </span>
                </button>
              </div>
              
              <div className="relative h-64 md:h-80 flex items-center justify-center">
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="w-64 h-64 rounded-full bg-gradient-to-br from-blue-600/20 to-purple-600/20 animate-pulse" />
                </div>
                <div className="relative grid grid-cols-3 gap-4">
                  {[Server, Shield, BarChart, Cloud, Users, Cpu].map((Icon, index) => (
                    <div
                      key={index}
                      className="p-4 rounded-xl glass border border-white/10 hover:scale-110 transition-transform duration-300"
                      style={{ animationDelay: `${index * 100}ms` }}
                    >
                      <Icon className="w-8 h-8 text-white" />
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
