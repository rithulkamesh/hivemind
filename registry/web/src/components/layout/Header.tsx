import { Link, useNavigate } from "react-router-dom";
import { Search, LogIn, User, LogOut, LayoutDashboard, Github } from "lucide-react";
import { useState } from "react";
import { useSession, signOut, useMe } from "@/store/auth";
import { SearchBar } from "@/search/SearchBar";

export function Header() {
  const { data: session } = useSession();
  const { data: me } = useMe();
  const navigate = useNavigate();
  const [searchOpen, setSearchOpen] = useState(false);
  const displayName = me?.username ?? session?.user?.name ?? session?.user?.email ?? "Account";

  return (
    <header className="sticky top-0 z-50 bg-hm-bg/95 backdrop-blur border-b border-hm-border min-h-[48px] flex items-center px-hm-lg">
      <div className="w-full max-w-6xl mx-auto flex items-center justify-between gap-4">
        <Link
          to="/"
          className="font-mono text-xs tracking-widest uppercase text-hm-text hover:opacity-90 transition-opacity"
        >
          Hivemind Registry
        </Link>
        
        <Link 
          to="/packages" 
          className="ml-4 font-sans text-sm font-medium text-hm-muted hover:text-hm-text transition-colors"
        >
          Packages
        </Link>

        <div className="flex-1 max-w-xl mx-4 hidden sm:block">
          <SearchBar inline onSelect={() => setSearchOpen(false)} />
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            className="sm:hidden p-2 text-hm-muted hover:text-hm-text transition-opacity"
            onClick={() => setSearchOpen((o) => !o)}
            aria-label="Search"
          >
            <Search size={18} />
          </button>

          {session?.user ? (
            <>
              <Link
                to="/dashboard"
                className="p-2 text-hm-muted hover:text-hm-text transition-opacity"
                title="Dashboard"
              >
                <LayoutDashboard size={18} />
              </Link>
              <div className="relative group">
                <button
                  type="button"
                  className="flex items-center gap-2 px-2 py-1.5 text-hm-muted hover:text-hm-text transition-opacity font-sans text-sm"
                  aria-expanded="false"
                  aria-haspopup="true"
                >
                  <User size={18} />
                  <span className="hidden sm:inline">{displayName}</span>
                </button>
                <div className="absolute right-0 top-full mt-1 py-1 w-48 bg-hm-surface border border-hm-border rounded shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
                  <Link
                    to="/dashboard/settings"
                    className="block px-3 py-2 text-sm text-hm-text-passive hover:text-hm-text hover:bg-hm-border/50"
                  >
                    Settings
                  </Link>
                  <button
                    type="button"
                    className="w-full text-left px-3 py-2 text-sm text-hm-text-passive hover:text-hm-text hover:bg-hm-border/50 flex items-center gap-2"
                    onClick={() => {
                      signOut();
                      navigate("/");
                    }}
                  >
                    <LogOut size={14} />
                    Log out
                  </button>
                </div>
              </div>
            </>
          ) : (
            <Link
              to="/login"
              className="flex items-center gap-2 px-3 py-1.5 text-sm text-hm-muted hover:text-hm-text transition-opacity"
            >
              <LogIn size={18} />
              <span>Log in</span>
            </Link>
          )}

          <a
            href="https://github.com/rithulkamesh/hivemind"
            target="_blank"
            rel="noopener noreferrer"
            className="p-2 text-hm-muted hover:text-hm-text transition-opacity"
            aria-label="GitHub"
          >
            <Github size={18} />
          </a>
        </div>
      </div>

      {searchOpen && (
        <div className="sm:hidden absolute top-full left-0 right-0 p-2 bg-hm-bg border-b border-hm-border">
          <SearchBar inline onSelect={() => setSearchOpen(false)} />
        </div>
      )}
    </header>
  );
}
