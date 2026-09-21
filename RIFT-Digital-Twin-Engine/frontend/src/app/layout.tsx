import type { Metadata } from "next";
import "./globals.css";
import { RiftProviders } from "@/components/providers";
import { AuthProvider } from "@/components/auth-provider";

export const metadata: Metadata = {
  title: "RIFT — Digital Twin Engine",
  description: "Universal Digital Twin Operating Platform",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" dir="ltr" data-theme="windows" suppressHydrationWarning>
      <body suppressHydrationWarning>
        <RiftProviders>
          <AuthProvider>{children}</AuthProvider>
        </RiftProviders>
      </body>
    </html>
  );
}
