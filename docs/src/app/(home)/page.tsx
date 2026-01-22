import Link from 'next/link';

export default function HomePage() {
  return (
    <main className="flex flex-col flex-1 px-4 py-12 md:py-24 relative overflow-hidden">
      {/* Background Orbs */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full -z-10 pointer-events-none opacity-20">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-indigo-500 rounded-full blur-[120px] animate-pulse"></div>
        <div className="absolute bottom-[20%] right-[-5%] w-[35%] h-[35%] bg-purple-500 rounded-full blur-[100px] animate-pulse delay-700"></div>
        <div className="absolute top-[40%] left-[20%] w-[25%] h-[25%] bg-cyan-400 rounded-full blur-[80px] animate-pulse delay-1000"></div>
      </div>

      <div className="max-w-6xl mx-auto flex flex-col items-center text-center">
        {/* Badge */}
        <div className="mb-6 px-4 py-1.5 rounded-full bg-white/5 border border-white/10 backdrop-blur-md inline-flex items-center gap-2">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
          </span>
          <span className="text-sm font-medium text-indigo-300">OpenAI compatible, cost-aware routing</span>
        </div>

        {/* Hero Title */}
        <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-8 bg-gradient-to-br from-white via-indigo-100 to-white/40 bg-clip-text text-transparent leading-[1.1]">
          The Smart Gateway <br /> for LLM Infrastructure
        </h1>

        <p className="max-w-2xl text-lg md:text-xl text-zinc-400 mb-10 leading-relaxed">
          Route requests based on intent, cost, and latency.
          Maximize resilience with automatic fallbacks and local embedding-based classification.
        </p>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row gap-4 mb-20">
          <Link
            href="/docs"
            className="px-8 py-3.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-all shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_25px_rgba(79,70,229,0.5)] active:scale-95"
          >
            Get Started
          </Link>
          <Link
            href="https://github.com/oviecodes/octo-router"
            className="px-8 py-3.5 rounded-xl bg-white/5 hover:bg-white/10 text-white font-semibold border border-white/10 transition-all backdrop-blur-sm active:scale-95"
          >
            View on GitHub
          </Link>
        </div>

        {/* Feature Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left">
          <FeatureCard
            title="Semantic Routing"
            description="Locally run ONNX embeddings to classify intent and route to the best model for the task."
            icon="🧠"
          />
          <FeatureCard
            title="Cost Performance"
            description="Manage budgets with Redis-backed state. Automatically swap models to stay under limit."
            icon="💰"
          />
          <FeatureCard
            title="Cloud Native"
            description="Built in Go, ready for Docker. Compatible with Kubernetes and modern orchestrators."
            icon="🐳"
          />
        </div>

        {/* Preview Section */}
        <div className="mt-20 w-full max-w-4xl rounded-2xl border border-white/10 bg-black/40 backdrop-blur-xl overflow-hidden shadow-2xl">
          <div className="bg-white/5 px-4 py-2 flex items-center gap-1.5 border-b border-white/10">
            <div className="w-3 h-3 rounded-full bg-red-500/50"></div>
            <div className="w-3 h-3 rounded-full bg-yellow-500/50"></div>
            <div className="w-3 h-3 rounded-full bg-green-500/50"></div>
            <span className="ml-2 text-xs text-zinc-500 font-mono italic">config.yaml</span>
          </div>
          <div className="p-6 overflow-x-auto text-left">
            <pre className="text-sm font-mono text-zinc-300">
              <code>{`routing:
  strategy: "weighted"
  policies:
    semantic:
      enabled: true
      engine: "embedding"
      model_path: "assets/models/embedding.onnx"
      groups:
        - name: "coding"
          required_capability: "code-gen"
          allow_providers: ["openai", "anthropic"]`}</code>
            </pre>
          </div>
        </div>
      </div>
    </main>
  );
}

function FeatureCard({ title, description, icon }: { title: string; description: string; icon: string }) {
  return (
    <div className="p-6 rounded-2xl bg-white/[0.03] border border-white/5 hover:border-indigo-500/30 hover:bg-white/[0.05] transition-all group">
      <div className="text-3xl mb-4 grayscale group-hover:grayscale-0 transition-all duration-500">{icon}</div>
      <h3 className="text-lg font-bold text-white mb-2">{title}</h3>
      <p className="text-zinc-400 text-sm leading-relaxed">{description}</p>
    </div>
  );
}
