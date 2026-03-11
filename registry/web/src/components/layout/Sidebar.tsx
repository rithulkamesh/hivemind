import { Link, useParams, useLocation } from "react-router-dom";
import { clsx } from "clsx";

const dashboardNav = [
  { to: "/dashboard", label: "Overview" },
  { to: "/dashboard/packages", label: "Packages" },
  { to: "/dashboard/api-keys", label: "API keys" },
  { to: "/dashboard/settings", label: "Settings" },
];

const orgNav = (slug: string) => [
  { to: `/org/${slug}`, label: "Overview" },
  { to: `/org/${slug}/members`, label: "Members" },
  { to: `/org/${slug}/sso`, label: "SSO" },
  { to: `/org/${slug}/packages`, label: "Packages" },
];

const adminNav = [
  { to: "/admin/queue", label: "Verification queue" },
  { to: "/admin/users", label: "Users" },
  { to: "/admin/packages", label: "Packages" },
];

type SidebarVariant = "dashboard" | "org" | "admin";

interface SidebarProps {
  variant: SidebarVariant;
  className?: string;
}

export function Sidebar({ variant, className }: SidebarProps) {
  const { slug } = useParams<{ slug: string }>();
  const location = useLocation();
  const pathname = location.pathname;
  const items =
    variant === "dashboard"
      ? dashboardNav
      : variant === "org" && slug
        ? orgNav(slug)
        : adminNav;

  return (
    <aside
      className={clsx(
        "w-[200px] shrink-0 font-mono text-[9px] tracking-wider uppercase text-hm-muted",
        className
      )}
    >
      <nav className="flex flex-col gap-0">
        {items.map(({ to, label }) => (
          <Link
            key={to}
            to={to}
            className={clsx(
              "py-hm-md px-hm-md border-l-2 border-transparent transition-colors",
              "hover:text-hm-text",
              pathname === to || (to !== "/dashboard" && pathname.startsWith(to + "/"))
                ? "text-hm-text border-hm-text font-medium"
                : ""
            )}
          >
            {label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
