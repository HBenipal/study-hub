import { Toaster } from "@/components/ui/sonner";
import "./globals.css";
import { TooltipProvider } from "@/components/ui/tooltip";

export const metadata = {
  title: "Study Hub",
  description: "Realtime collaborative study hub",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          rel="preconnect"
          href="https://fonts.gstatic.com"
          crossOrigin="anonymous"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=Roboto:wght@300;400;500;600;700&display=swap"
          rel="stylesheet"
        />
      </head>

      <body>
        <TooltipProvider delayDuration={100}>
          {children} <Toaster position="top-right" richColors theme="light" />
        </TooltipProvider>
      </body>
    </html>
  );
}
