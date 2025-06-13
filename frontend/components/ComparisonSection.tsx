'use client'

import React, { useState } from 'react'
import { X, CheckCircle, XCircle, AlertCircle, ArrowRight, Clock, Code2, Users, Zap, Layers } from 'lucide-react'

const comparisonData = {
  traditional: {
    title: "Traditional Development",
    icon: Code2,
    color: "from-gray-500 to-gray-600",
    time: "2-6 months",
    cost: "$50,000 - $500,000",
    points: [
      { status: 'bad', text: "Manual coding for every component" },
      { status: 'bad', text: "Slow iteration cycles" },
      { status: 'bad', text: "High risk of technical debt" },
      { status: 'warning', text: "Requires large development team" },
      { status: 'warning', text: "Difficult to maintain consistency" },
      { status: 'bad', text: "Testing often an afterthought" },
    ],
    metrics: {
      speed: 20,
      quality: 60,
      cost: 10,
      scalability: 50,
    }
  },
  templates: {
    title: "Template-Based Generators",
    icon: Layers,
    color: "from-yellow-500 to-orange-500",
    time: "1-2 hours",
    cost: "$99 - $999/month",
    points: [
      { status: 'warning', text: "Limited to predefined patterns" },
      { status: 'good', text: "Quick initial setup" },
      { status: 'bad', text: "Generic, non-optimized output" },
      { status: 'warning', text: "Difficult to customize" },
      { status: 'bad', text: "One-size-fits-all approach" },
      { status: 'warning', text: "Basic documentation only" },
    ],
    metrics: {
      speed: 80,
      quality: 40,
      cost: 70,
      scalability: 30,
    }
  },
  quantumlayer: {
    title: "QuantumLayer LLOM",
    icon: Zap,
    color: "from-blue-500 to-purple-600",
    time: "12.5 seconds",
    cost: "From $79/month",
    points: [
      { status: 'good', text: "AI orchestrates specialized agents" },
      { status: 'good', text: "Unique code for every project" },
      { status: 'good', text: "Enterprise-grade quality" },
      { status: 'good', text: "Complete test coverage included" },
      { status: 'good', text: "Production-ready from day one" },
      { status: 'good', text: "Continuous learning & improvement" },
    ],
    metrics: {
      speed: 100,
      quality: 94,
      cost: 90,
      scalability: 95,
    }
  }
}

export default function ComparisonSection() {
  const [selectedMethod, setSelectedMethod] = useState<keyof typeof comparisonData>('quantumlayer')

  return (
    <section className="py-32 relative overflow-hidden">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-5">
        <div className="absolute inset-0" style={{
          backgroundImage: `linear-gradient(30deg, transparent 49%, white 49%, white 51%, transparent 51%)`,
          backgroundSize: '20px 20px'
        }} />
      </div>

      <div className="container mx-auto px-6 relative z-10">
        <div className="max-w-6xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-16">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass border border-white/10 mb-6">
              <AlertCircle className="w-4 h-4 text-yellow-400" />
              <span className="text-sm font-medium">Why QuantumLayer?</span>
            </div>
            
            <h2 className="text-4xl md:text-5xl font-bold mb-6">
              The Evolution of{' '}
              <span className="gradient-text">Software Development</span>
            </h2>
            
            <p className="text-xl text-gray-400 max-w-3xl mx-auto">
              See how QuantumLayer compares to traditional development methods and template-based generators
            </p>
          </div>

          {/* Comparison Cards */}
          <div className="grid md:grid-cols-3 gap-6 mb-16">
            {Object.entries(comparisonData).map(([key, data]) => {
              const Icon = data.icon
              const isSelected = selectedMethod === key
              
              return (
                <button
                  key={key}
                  onClick={() => setSelectedMethod(key as keyof typeof comparisonData)}
                  className={`glass rounded-2xl p-6 border transition-all duration-300 text-left ${
                    isSelected
                      ? 'border-white/20 scale-105 shadow-2xl'
                      : 'border-white/5 hover:border-white/10'
                  }`}
                >
                  <div className="flex items-center gap-4 mb-4">
                    <div className={`p-3 rounded-xl bg-gradient-to-br ${data.color}`}>
                      <Icon className="w-6 h-6 text-white" />
                    </div>
                    <h3 className="text-lg font-semibold flex-1">{data.title}</h3>
                    {key === 'quantumlayer' && (
                      <span className="px-2 py-1 text-xs bg-green-500/20 text-green-400 rounded-full">
                        Recommended
                      </span>
                    )}
                  </div>
                  
                  <div className="space-y-3 mb-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-400">Time to Build</span>
                      <span className="text-sm font-semibold">{data.time}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-400">Typical Cost</span>
                      <span className="text-sm font-semibold">{data.cost}</span>
                    </div>
                  </div>
                  
                  {/* Metrics Bars */}
                  <div className="space-y-2">
                    {Object.entries(data.metrics).map(([metric, value]) => (
                      <div key={metric}>
                        <div className="flex items-center justify-between text-xs mb-1">
                          <span className="capitalize text-gray-500">{metric}</span>
                          <span className="text-gray-400">{value}%</span>
                        </div>
                        <div className="h-1.5 bg-white/10 rounded-full overflow-hidden">
                          <div
                            className={`h-full bg-gradient-to-r ${data.color} transition-all duration-1000`}
                            style={{ width: isSelected ? `${value}%` : '0%' }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                </button>
              )
            })}
          </div>

          {/* Detailed Comparison */}
          <div className="glass rounded-3xl p-8 md:p-10 border border-white/10">
            <h3 className="text-2xl font-bold mb-8 text-center">
              Detailed Comparison: {comparisonData[selectedMethod].title}
            </h3>
            
            <div className="grid md:grid-cols-2 gap-8">
              {/* Features List */}
              <div>
                <h4 className="font-semibold mb-4 flex items-center gap-2">
                  <CheckCircle className="w-5 h-5 text-green-400" />
                  Key Characteristics
                </h4>
                
                <div className="space-y-3">
                  {comparisonData[selectedMethod].points.map((point, index) => (
                    <div key={index} className="flex items-start gap-3">
                      {point.status === 'good' && (
                        <CheckCircle className="w-5 h-5 text-green-400 mt-0.5 flex-shrink-0" />
                      )}
                      {point.status === 'bad' && (
                        <XCircle className="w-5 h-5 text-red-400 mt-0.5 flex-shrink-0" />
                      )}
                      {point.status === 'warning' && (
                        <AlertCircle className="w-5 h-5 text-yellow-400 mt-0.5 flex-shrink-0" />
                      )}
                      <span className="text-sm text-gray-300">{point.text}</span>
                    </div>
                  ))}
                </div>
              </div>
              
              {/* Visual Comparison */}
              <div>
                <h4 className="font-semibold mb-4 flex items-center gap-2">
                  <Zap className="w-5 h-5 text-purple-400" />
                  Performance Metrics
                </h4>
                
                <div className="relative h-64 flex items-end justify-between gap-4">
                  {Object.entries(comparisonData[selectedMethod].metrics).map(([metric, value]) => (
                    <div key={metric} className="flex-1 flex flex-col items-center">
                      <div className="relative w-full flex flex-col items-center">
                        <div
                          className={`w-full bg-gradient-to-t ${comparisonData[selectedMethod].color} rounded-t-lg transition-all duration-1000`}
                          style={{ height: `${(value / 100) * 200}px` }}
                        />
                        <span className="text-2xl font-bold mt-2">{value}%</span>
                      </div>
                      <span className="text-xs text-gray-400 capitalize mt-2">{metric}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
            
            {/* Call to Action */}
            {selectedMethod === 'quantumlayer' && (
              <div className="mt-8 p-6 rounded-2xl bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-semibold mb-2">Ready to experience the future?</h4>
                    <p className="text-sm text-gray-400">
                      Join thousands of developers building with QuantumLayer
                    </p>
                  </div>
                  <button className="group px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl font-semibold hover:shadow-lg hover:shadow-purple-500/25 transition-all duration-300">
                    <span className="flex items-center gap-2">
                      Start Building
                      <ArrowRight className="w-4 h-4 transition-transform duration-300 group-hover:translate-x-1" />
                    </span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  )
}
