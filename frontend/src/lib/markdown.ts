import DOMPurify from "dompurify";
import MarkdownIt from "markdown-it";

const markdownRenderer = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true,
});

interface RenderSimpleMarkdownOptions {
  resolveURL?: (rawURL: string, tagName: "a" | "img") => string | null;
}

export function renderSimpleMarkdown(
  source: string,
  options: RenderSimpleMarkdownOptions = {},
) {
  const normalized = source.replace(/\r\n/g, "\n").trim();
  if (!normalized) {
    return "";
  }

  const rendered = markdownRenderer.render(normalized);
  const sanitized = DOMPurify.sanitize(rendered, {
    USE_PROFILES: { html: true },
  });
  if (!options.resolveURL) {
    return sanitized;
  }

  const transformed = rewriteRelativeMarkdownURLs(sanitized, options.resolveURL);
  return DOMPurify.sanitize(transformed, {
    USE_PROFILES: { html: true },
  });
}

function rewriteRelativeMarkdownURLs(
  html: string,
  resolveURL: NonNullable<RenderSimpleMarkdownOptions["resolveURL"]>,
) {
  const documentFragment = new DOMParser().parseFromString(html, "text/html");

  for (const element of documentFragment.querySelectorAll("a[href], img[src]")) {
    if (element instanceof HTMLAnchorElement) {
      rewriteElementURL(element, "href", "a", resolveURL);
      continue;
    }
    if (element instanceof HTMLImageElement) {
      rewriteElementURL(element, "src", "img", resolveURL);
    }
  }

  return documentFragment.body.innerHTML;
}

function rewriteElementURL(
  element: HTMLAnchorElement | HTMLImageElement,
  attributeName: "href" | "src",
  tagName: "a" | "img",
  resolveURL: NonNullable<RenderSimpleMarkdownOptions["resolveURL"]>,
) {
  const rawURL = element.getAttribute(attributeName)?.trim();
  if (!rawURL || !isRelativeMarkdownURL(rawURL)) {
    return;
  }

  const nextURL = resolveURL(rawURL, tagName);
  if (!nextURL) {
    return;
  }

  element.setAttribute(attributeName, nextURL);
}

function isRelativeMarkdownURL(rawURL: string) {
  if (!rawURL || rawURL.startsWith("#") || rawURL.startsWith("/")) {
    return false;
  }
  if (rawURL.startsWith("//")) {
    return false;
  }

  return !/^[a-zA-Z][a-zA-Z\d+.-]*:/.test(rawURL);
}
