export const QuantumLayerLogo = ({ className = "w-10 h-10" }: { className?: string }) => (
  <svg 
    className={className} 
    viewBox="0 0 100 100" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="quantum-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#0EA5E9" />
        <stop offset="50%" stopColor="#8B5CF6" />
        <stop offset="100%" stopColor="#06B6D4" />
      </linearGradient>
      
      <filter id="glow">
        <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>
    
    {/* Outer quantum ring */}
    <circle 
      cx="50" 
      cy="50" 
      r="45" 
      stroke="url(#quantum-gradient)" 
      strokeWidth="2" 
      fill="none"
      opacity="0.3"
    />
    
    {/* Middle quantum ring */}
    <circle 
      cx="50" 
      cy="50" 
      r="35" 
      stroke="url(#quantum-gradient)" 
      strokeWidth="1.5" 
      fill="none"
      opacity="0.5"
      strokeDasharray="10 5"
    >
      <animateTransform
        attributeName="transform"
        attributeType="XML"
        type="rotate"
        from="0 50 50"
        to="360 50 50"
        dur="20s"
        repeatCount="indefinite"
      />
    </circle>
    
    {/* Inner quantum ring */}
    <circle 
      cx="50" 
      cy="50" 
      r="25" 
      stroke="url(#quantum-gradient)" 
      strokeWidth="1" 
      fill="none"
      opacity="0.7"
      strokeDasharray="5 10"
    >
      <animateTransform
        attributeName="transform"
        attributeType="XML"
        type="rotate"
        from="360 50 50"
        to="0 50 50"
        dur="15s"
        repeatCount="indefinite"
      />
    </circle>
    
    {/* Core hexagon */}
    <path
      d="M50 20 L70 35 L70 65 L50 80 L30 65 L30 35 Z"
      fill="url(#quantum-gradient)"
      filter="url(#glow)"
    />
    
    {/* Q Letter */}
    <text 
      x="50" 
      y="58" 
      fontFamily="Arial, sans-serif" 
      fontSize="28" 
      fontWeight="bold" 
      fill="white" 
      textAnchor="middle"
    >
      Q
    </text>
    
    {/* Quantum particles */}
    <circle cx="20" cy="20" r="2" fill="#0EA5E9">
      <animate 
        attributeName="opacity" 
        values="0;1;0" 
        dur="3s" 
        repeatCount="indefinite" 
      />
    </circle>
    
    <circle cx="80" cy="20" r="2" fill="#8B5CF6">
      <animate 
        attributeName="opacity" 
        values="0;1;0" 
        dur="3s" 
        begin="1s"
        repeatCount="indefinite" 
      />
    </circle>
    
    <circle cx="80" cy="80" r="2" fill="#06B6D4">
      <animate 
        attributeName="opacity" 
        values="0;1;0" 
        dur="3s" 
        begin="2s"
        repeatCount="indefinite" 
      />
    </circle>
    
    <circle cx="20" cy="80" r="2" fill="#6366F1">
      <animate 
        attributeName="opacity" 
        values="0;1;0" 
        dur="3s" 
        begin="1.5s"
        repeatCount="indefinite" 
      />
    </circle>
  </svg>
)

export const QuantumLayerLogomark = ({ className = "w-8 h-8" }: { className?: string }) => (
  <svg 
    className={className} 
    viewBox="0 0 40 40" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <linearGradient id="logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#0EA5E9" />
        <stop offset="100%" stopColor="#8B5CF6" />
      </linearGradient>
    </defs>
    
    {/* Rounded square background */}
    <rect 
      width="40" 
      height="40" 
      rx="10" 
      fill="url(#logo-gradient)"
    />
    
    {/* Inner dark square */}
    <rect 
      x="4" 
      y="4" 
      width="32" 
      height="32" 
      rx="8" 
      fill="#0A0A0B"
    />
    
    {/* Q Letter */}
    <text 
      x="20" 
      y="26" 
      fontFamily="Arial, sans-serif" 
      fontSize="20" 
      fontWeight="bold" 
      fill="url(#logo-gradient)" 
      textAnchor="middle"
    >
      Q
    </text>
    
    {/* Corner accent */}
    <circle 
      cx="32" 
      cy="8" 
      r="3" 
      fill="#06B6D4"
    />
  </svg>
)
