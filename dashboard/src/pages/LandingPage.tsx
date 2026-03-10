import { useEffect } from 'react';
import Navbar from '../components/landing/Navbar';
import Hero from '../components/landing/Hero';
import LogoCloud from '../components/landing/LogoCloud';
import Problem from '../components/landing/Problem';
import HowItWorks from '../components/landing/HowItWorks';
import Features from '../components/landing/Features';
import Pricing from '../components/landing/Pricing';
import Stats from '../components/landing/Stats';
import MidPageCTA from '../components/landing/MidPageCTA';
import CodeExample from '../components/landing/CodeExample';
import UseCases from '../components/landing/UseCases';
import Comparison from '../components/landing/Comparison';
import GetStarted from '../components/landing/GetStarted';
import FAQ from '../components/landing/FAQ';
import OpenSourceCTA from '../components/landing/OpenSourceCTA';
import Footer from '../components/landing/Footer';

export default function LandingPage() {
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) =>
        entries.forEach((e) => {
          if (e.isIntersecting) e.target.classList.add('visible');
        }),
      { threshold: 0.1 }
    );

    document.querySelectorAll('.scroll-reveal').forEach((el) => observer.observe(el));

    return () => observer.disconnect();
  }, []);

  return (
    <main className="bg-gray-950 min-h-screen text-white">
      <Navbar />
      <Hero />
      <LogoCloud />
      <Problem />
      <HowItWorks />
      <Features />
      <Pricing />
      <Stats />
      <MidPageCTA />
      <CodeExample />
      <UseCases />
      <Comparison />
      <GetStarted />
      <FAQ />
      <OpenSourceCTA />
      <Footer />
    </main>
  );
}
