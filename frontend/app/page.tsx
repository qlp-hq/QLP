import HeroSection from '@/components/HeroSection'
import ArchitectureDiagram from '@/components/ArchitectureDiagram'
import Playground from '@/components/Playground'
import ComparisonSection from '@/components/ComparisonSection'
import IntelligencePipeline from '@/components/IntelligencePipeline'
import AdvancedFeatures from '@/components/AdvancedFeatures'
import Navigation from '@/components/Navigation'
import AIAssistant from '@/components/AIAssistant'
import FloatingParticles from '@/components/FloatingParticles'
import Footer from '@/components/Footer'

export default function Home() {
  return (
    <>
      <FloatingParticles />
      <div className="relative z-10">
        <Navigation />
        <main className="min-h-screen">
          <HeroSection />
          <Playground />
          <ArchitectureDiagram />
          <IntelligencePipeline />
          <ComparisonSection />
          <AdvancedFeatures />
        </main>
        <Footer />
        <AIAssistant />
      </div>
    </>
  )
}
