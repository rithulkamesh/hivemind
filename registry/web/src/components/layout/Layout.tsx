import { Outlet } from "react-router-dom";
import { Header } from "./Header";
import { Footer } from "./Footer";

export function Layout() {
  return (
    <div className="min-h-screen flex flex-col bg-hm-bg text-hm-text font-sans">
      <Header />
      <main className="flex-1 w-full max-w-6xl mx-auto px-hm-lg py-hm-xl box-border">
        <Outlet />
      </main>
      <Footer />
    </div>
  );
}
