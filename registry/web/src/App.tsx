import { Routes, Route, Navigate } from "react-router-dom";
import { Layout } from "@/components/layout/Layout";
import { Home } from "@/pages/Home";
import { Packages } from "@/pages/Packages";
import { Search } from "@/pages/Search";
import { PackageDetail } from "@/pages/PackageDetail";
import { PackageVersion } from "@/pages/PackageVersion";
import { PublishGuide } from "@/pages/PublishGuide";
import { Login } from "@/pages/Login";
import { Register } from "@/pages/Register";
import { VerifyEmail } from "@/pages/VerifyEmail";
import { ErrorPage } from "@/pages/ErrorPage";
import { Dashboard } from "@/pages/Dashboard";
import { DashboardPackages } from "@/pages/Dashboard/Packages";
import { CreatePackage } from "@/pages/Dashboard/CreatePackage";
import { ApiKeys } from "@/pages/Dashboard/ApiKeys";
import { Settings } from "@/pages/Dashboard/Settings";
import { OrgIndex } from "@/pages/Org";
import { OrgMembers } from "@/pages/Org/Members";
import { OrgSSO } from "@/pages/Org/SSO";
import { OrgPackages } from "@/pages/Org/Packages";
import { AdminQueue } from "@/pages/Admin/Queue";
import { AdminUsers } from "@/pages/Admin/Users";
import { AdminPackages } from "@/pages/Admin/Packages";
import { AuthGuard } from "@/components/auth/AuthGuard";
import { GuestGuard } from "@/components/auth/GuestGuard";
import { OrgGuard } from "@/components/auth/OrgGuard";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Home />} />
        <Route path="packages" element={<Packages />} />
        <Route path="search" element={<Search />} />
        <Route path="packages/:name" element={<PackageDetail />} />
        <Route path="packages/:name/v/:version" element={<PackageVersion />} />
        <Route path="publish" element={<PublishGuide />} />
        <Route path="login" element={<GuestGuard><Login /></GuestGuard>} />
        <Route path="register" element={<GuestGuard><Register /></GuestGuard>} />
        <Route path="verify-email" element={<VerifyEmail />} />
        <Route path="error" element={<ErrorPage />} />

        <Route path="dashboard" element={<AuthGuard><Dashboard /></AuthGuard>} />
        <Route path="dashboard/packages" element={<AuthGuard><DashboardPackages /></AuthGuard>} />
        <Route path="dashboard/packages/new" element={<AuthGuard><CreatePackage /></AuthGuard>} />
        <Route path="dashboard/api-keys" element={<AuthGuard><ApiKeys /></AuthGuard>} />
        <Route path="dashboard/settings" element={<AuthGuard><Settings /></AuthGuard>} />

        <Route path="org/:slug" element={<AuthGuard><OrgIndex /></AuthGuard>} />
        <Route path="org/:slug/members" element={<AuthGuard><OrgGuard><OrgMembers /></OrgGuard></AuthGuard>} />
        <Route path="org/:slug/sso" element={<AuthGuard><OrgGuard><OrgSSO /></OrgGuard></AuthGuard>} />
        <Route path="org/:slug/packages" element={<AuthGuard><OrgGuard><OrgPackages /></OrgGuard></AuthGuard>} />

        <Route path="admin/queue" element={<AuthGuard><AdminQueue /></AuthGuard>} />
        <Route path="admin/users" element={<AuthGuard><AdminUsers /></AuthGuard>} />
        <Route path="admin/packages" element={<AuthGuard><AdminPackages /></AuthGuard>} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
