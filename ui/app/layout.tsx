import type { Metadata } from "next";
import { Inter, Roboto_Mono } from "next/font/google"; // Use nice fonts
import "./globals.css";
import Sidebar from "@/components/layout/Sidebar";
import Header from "@/components/layout/Header";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const robotoMono = Roboto_Mono({ subsets: ["latin"], variable: "--font-roboto-mono" });

export const metadata: Metadata = {
  title: "PhishVault 2.0 | Analyst Workbench",
  description: "Advanced Phishing Intelligence Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${robotoMono.variable} antialiased bg-slate-950 text-slate-200`}>
        <div className="flex min-h-screen">
          <Sidebar />
          <div className="flex-1 ml-64">
            <Header />
            <main className="p-6 mt-16">
              {children}
            </main>
          </div>
        </div>
      </body>
    </html>
  );
}
