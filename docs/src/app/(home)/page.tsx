import Link from 'next/link';

export default function HomePage() {
  return (
    <main className="flex flex-col flex-1 relative overflow-hidden bg-white dark:bg-black transition-colors duration-300">
      {/* Background Orbs */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full -z-10 pointer-events-none opacity-20 dk:opacity-20">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-indigo-500 rounded-full blur-[120px] animate-pulse"></div>
        <div className="absolute bottom-[20%] right-[-5%] w-[35%] h-[35%] bg-purple-500 rounded-full blur-[100px] animate-pulse delay-700"></div>
        <div className="absolute top-[40%] left-[20%] w-[25%] h-[25%] bg-cyan-400 rounded-full blur-[80px] animate-pulse delay-1000"></div>
      </div>

      {/* HERO SECTION */}
      <section className="flex flex-col items-center text-center px-4 py-16 md:py-32 max-w-6xl mx-auto">
        {/* Badge */}
        <div className="mb-8 px-4 py-1.5 rounded-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 backdrop-blur-md inline-flex items-center gap-2">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
          </span>
          <span className="text-sm font-medium text-slate-600 dark:text-indigo-300">Production-ready AI Gateway</span>
        </div>

        {/* Hero Title */}
        <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-8 text-slate-900 dark:text-transparent dark:bg-clip-text dark:bg-gradient-to-br dark:from-white dark:via-indigo-100 dark:to-zinc-400 leading-[1.1]">
          The Smart Gateway <br /> for LLM Infrastructure
        </h1>

        <p className="max-w-2xl text-lg md:text-xl text-slate-600 dark:text-zinc-400 mb-12 leading-relaxed">
          Route requests based on <span className="text-indigo-600 dark:text-indigo-400 font-semibold">intent</span>, <span className="text-indigo-600 dark:text-indigo-400 font-semibold">cost</span>, and <span className="text-indigo-600 dark:text-indigo-400 font-semibold">latency</span>.
          Maximize resilience with automatic fallbacks and local embedding-based classification.
        </p>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row gap-4 mb-20 animate-in fade-in slide-in-from-bottom-4 duration-1000 fill-mode-forwards">
          <Link
            href="/docs"
            className="px-8 py-3.5 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white font-semibold transition-all shadow-lg hover:shadow-indigo-500/30 active:scale-95 flex items-center justify-center"
          >
            Get Started <span className="ml-2">→</span>
          </Link>
          <Link
            href="https://github.com/oviecodes/octo-router"
            className="px-8 py-3.5 rounded-xl bg-white dark:bg-white/5 hover:bg-slate-50 dark:hover:bg-white/10 text-slate-900 dark:text-white font-semibold border border-slate-200 dark:border-white/10 transition-all backdrop-blur-sm active:scale-95 flex items-center justify-center"
          >
            View on GitHub
          </Link>
        </div>

        {/* Feature Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left w-full">
          <FeatureCard
            title="Semantic Routing"
            description="Locally run ONNX embeddings to classify intent. Route 'coding' to Claude 3.5 Sonnet and 'chit-chat' to GPT-4o-mini automatically."
            icon="🧠"
          />
          <FeatureCard
            title="Cost Controls"
            description="Set daily budgets per provider. Automatically skip providers that exceed limits to prevent overspending and ensure budget predictability."
            icon="💰"
          />
          <FeatureCard
            title="Docker Ready"
            description="Fully containerized architecture. Deploy instantly with Docker Compose and scale horizontally thanks to stateless Redis-backed tracking."
            icon="🐳"
          />
        </div>
      </section>

      {/* HOW IT WORKS SECTION */}
      <section className="py-24 bg-slate-50/50 dark:bg-white/[0.02] border-y border-slate-200 dark:border-white/5">
        <div className="max-w-5xl mx-auto px-4">
          <h2 className="text-3xl md:text-4xl font-bold text-center mb-16 text-slate-900 dark:text-white">How Octo Router Works</h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 relative">
            {/* Connector Lines (Desktop) */}
            <div className="hidden md:block absolute top-12 left-[16%] w-[68%] h-0.5 bg-gradient-to-r from-transparent via-indigo-300 dark:via-indigo-500/30 to-transparent"></div>

            <StepCard
              number="1"
              title="Receive Request"
              description="Your app sends a standard OpenAI-compatible chat completion request."
            />
            <StepCard
              number="2"
              title="Route & Optimize"
              description="Octo Router classifies intent, checks budgets, and selects the best provider."
            />
            <StepCard
              number="3"
              title="Proxy Response"
              description="The request is forwarded, and the response is streamed back with usage stats."
            />
          </div>
        </div>
      </section>

      {/* CODE PREVIEW SECTION */}
      <section className="py-24 px-4">
        <div className="max-w-5xl mx-auto flex flex-col md:flex-row items-center gap-12">
          <div className="flex-1 text-center md:text-left">
            <h2 className="text-3xl md:text-4xl font-bold mb-6 text-slate-900 dark:text-white">
              Configuration as Code
            </h2>
            <p className="text-lg text-slate-600 dark:text-zinc-400 mb-8">
              Define your routing logic, budgets, and fallbacks in a simple, clear YAML file. No complex UIs or proprietary databases required.
            </p>
            <ul className="space-y-4 text-left inline-block">
              {[
                "Git-ops friendly configuration",
                "Hot-reloading without downtime",
                "Environment variable substitution"
              ].map((item, i) => (
                <li key={i} className="flex items-center gap-3 text-slate-700 dark:text-zinc-300">
                  <svg className="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
                  {item}
                </li>
              ))}
            </ul>
          </div>

          <div className="flex-1 w-full max-w-lg rounded-2xl border border-slate-200 dark:border-white/10 bg-white dark:bg-black/40 shadow-2xl overflow-hidden">
            <div className="bg-slate-100 dark:bg-white/5 px-4 py-2 flex items-center gap-1.5 border-b border-slate-200 dark:border-white/10">
              <div className="w-3 h-3 rounded-full bg-red-400"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
              <div className="w-3 h-3 rounded-full bg-green-400"></div>
              <span className="ml-2 text-xs text-slate-500 dark:text-zinc-500 font-mono italic">config.yaml</span>
            </div>
            <div className="p-6 overflow-x-auto text-left bg-slate-50 dark:bg-transparent">
              <pre className="text-sm font-mono text-slate-800 dark:text-zinc-300">
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
      </section>
      <section className="py-24 bg-slate-50 dark:bg-black/20 border-t border-slate-200 dark:border-white/5">
        <div className="max-w-6xl mx-auto px-4 text-center">
          <h2 className="text-3xl md:text-4xl font-bold mb-16 text-slate-900 dark:text-white">Built for Scalable AI Teams</h2>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            <UseCaseCard
              title="SaaS Tiered Pricing"
              description="Route requests to different models based on user priority or tier. Enforce strict budget limits per provider to maintain healthy margins."
            />
            <UseCaseCard
              title="Internal Developer Platform"
              description="Give every engineer a unified API endpoint. Monitor usage by team and prevent accidental overspending."
            />
            <UseCaseCard
              title="High-Traffic Consumer App"
              description="Monitor performance metrics in real-time. Use weighted routing to distribute load across multiple providers ensuring high availability."
            />
          </div>
        </div>
      </section>

      <section className="py-24 px-4">
        <div className="max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold mb-12 text-center text-slate-900 dark:text-white">Frequently Asked Questions</h2>

          <div className="space-y-6">
            <FAQItem
              question="Does Octo Router add latency?"
              answer="Minimal overhead (~2ms). It's written in Go and uses a highly optimized routing pipeline."
            />
            <FAQItem
              question="Can I use it with any LLM?"
              answer="Yes, as long as it has an OpenAI-compatible API (which most do, including vLLM, Ollama, and Groq)."
            />
            <FAQItem
              question="How does semantic routing work locally?"
              answer="We embed the ONNX runtime directly in the binary. This allows us to run quantization-optimized models (like all-MiniLM-L6-v2) on the CPU in milliseconds."
            />
          </div>
        </div>
      </section>
    </main>
  );
}

function UseCaseCard({ title, description }: { title: string; description: string }) {
  return (
    <div className="p-8 rounded-2xl bg-white dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 text-left hover:border-indigo-500/50 transition-colors">
      <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-3">{title}</h3>
      <p className="text-slate-600 dark:text-zinc-400 leading-relaxed">{description}</p>
    </div>
  );
}

function FAQItem({ question, answer }: { question: string; answer: string }) {
  return (
    <div className="p-6 rounded-2xl bg-slate-50 dark:bg-white/[0.03] border border-slate-200 dark:border-white/5">
      <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-2">{question}</h3>
      <p className="text-slate-600 dark:text-zinc-400">{answer}</p>
    </div>
  );
}

function FeatureCard({ title, description, icon }: { title: string; description: string; icon: string }) {
  return (
    <div className="p-6 rounded-2xl bg-white dark:bg-white/[0.03] border border-slate-200 dark:border-white/5 hover:border-indigo-500/30 hover:shadow-lg dark:hover:bg-white/[0.05] transition-all group">
      <div className="text-4xl mb-4 grayscale group-hover:grayscale-0 transition-all duration-500">{icon}</div>
      <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-2">{title}</h3>
      <p className="text-slate-600 dark:text-zinc-400 text-sm leading-relaxed">{description}</p>
    </div>
  );
}

function StepCard({ number, title, description }: { number: string; title: string; description: string }) {
  return (
    <div className="relative z-10 flex flex-col items-center text-center">
      <div className="w-12 h-12 rounded-full bg-indigo-600 text-white text-xl font-bold flex items-center justify-center mb-4 shadow-lg shadow-indigo-500/30">
        {number}
      </div>
      <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">{title}</h3>
      <p className="text-slate-600 dark:text-zinc-400 max-w-xs">{description}</p>
    </div>
  );
}
