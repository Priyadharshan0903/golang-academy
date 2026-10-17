import type { Metadata } from "next";
import "@/shared/styles/globals.css";

export const metadata: Metadata = {
  "title": "Go by Doing \u2014 GoPath Edition V2",
  "description": "116 GoPath lessons, 27 exercises, project sources, and company case studies.",
  "icons": {
    "icon": "/favicon.svg"
  }
};

export default function RootLayout({children}:Readonly<{children:React.ReactNode}>){return <html lang="en"><body>{children}</body></html>;}
