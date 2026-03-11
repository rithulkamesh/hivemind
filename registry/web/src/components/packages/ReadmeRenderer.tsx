import { useEffect } from "react";
import { marked } from "marked";
import Prism from "prismjs";
import "prismjs/themes/prism-tomorrow.css";

interface ReadmeRendererProps {
  content: string;
}

export function ReadmeRenderer({ content }: ReadmeRendererProps) {
  const html = marked(content, { gfm: true }) as string;

  useEffect(() => {
    Prism.highlightAll();
  }, [content]);

  return (
    <div
      className="prose-registry prose-registry"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
