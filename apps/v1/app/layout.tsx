import type { Metadata } from "next";
import "@/shared/styles/globals.css";

export const metadata: Metadata = {
  "title": "Go by Doing \u2014 Original Edition",
  "description": "48 practical Go lessons, exercises, quizzes, and the guided website-checker project.",
  "icons": {
    "icon": "/favicon.svg"
  }
};

export default function RootLayout({children}:Readonly<{children:React.ReactNode}>){return <html lang="en"><body>{children}</body></html>;}
