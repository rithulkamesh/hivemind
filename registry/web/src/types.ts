export interface User {
  id: string;
  email: string;
  username: string;
  email_verified?: boolean;
  totp_enabled?: boolean;
}

export interface Org {
  id: string;
  name: string;
  display_name: string;
}

export interface Package {
  id: string;
  name: string;
  namespace?: string | null;
  display_name: string;
  description?: string | null;
  homepage?: string | null;
  repository?: string | null;
  license?: string | null;
  keywords?: string[] | null;
  verified?: boolean;
  trusted?: boolean;
  total_downloads?: number;
  created_at?: string;
  owner_user_id?: string | null;
}

export interface PackageVersion {
  id: string;
  package_id: string;
  version: string;
  requires_python?: string | null;
  requires_hivemind?: string | null;
  tool_count?: number | null;
  uploaded_at?: string;
  published?: boolean;
  yanked?: boolean;
  yank_reason?: string | null;
  verification_status?: string | null;
  verification_report?: unknown;
}

export interface PackageFile {
  id: string;
  version_id: string;
  filename: string;
  filetype: string;
  size_bytes: number;
  sha256: string;
  md5: string;
  download_count?: number;
}

export interface Stats {
  total_packages: number;
  total_downloads: number;
}

export interface ApiKeyRow {
  id: string;
  name: string;
  key_prefix: string;
  scopes: string[];
  last_used_at?: string | null;
  expires_at?: string | null;
  created_at: string;
}

export interface ListPackagesResponse {
  packages: Package[];
  page: number;
}

export interface SearchResponse {
  results: Package[];
  page: number;
}
