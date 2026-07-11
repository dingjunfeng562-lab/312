import { blobEndpoint } from "@/conf/env.ts";
import { trimSuffixes } from "@/utils/base.ts";

export type BlobParserResponse = {
  status: boolean;
  content: string;
  error?: string;
};

const UploadTimeout = 2 * 60 * 1000;
export const ImageModelRequiredError = "IMAGE_MODEL_REQUIRED";

const imageMimeTypes: Record<string, string> = {
  png: "image/png",
  jpg: "image/jpeg",
  jpeg: "image/jpeg",
  gif: "image/gif",
  webp: "image/webp",
  bmp: "image/bmp",
  svg: "image/svg+xml",
  tif: "image/tiff",
  tiff: "image/tiff",
};

export type FileObject = {
  name: string;
  content: string;
  size?: number;
};

type Model = {
  id: string;
  ocr_model?: boolean;
  vision_model?: boolean;
  reverse_model?: boolean;
};

export type FileArray = FileObject[];

export const supportedFileExtensions = [
  // text and documents
  "txt",
  "md",
  "markdown",
  "rtf",
  "pdf",
  "doc",
  "docx",
  "odt",
  "ppt",
  "pptx",
  "odp",
  "xls",
  "xlsx",
  "ods",
  // images and audio
  "png",
  "jpg",
  "jpeg",
  "gif",
  "webp",
  "bmp",
  "svg",
  "tif",
  "tiff",
  "mp3",
  "wav",
  "m4a",
  "aac",
  "ogg",
  "oga",
  "flac",
  "opus",
  "webm",
  // code
  "c",
  "cc",
  "cpp",
  "cxx",
  "h",
  "hpp",
  "cs",
  "go",
  "java",
  "js",
  "jsx",
  "ts",
  "tsx",
  "py",
  "rb",
  "rs",
  "php",
  "swift",
  "kt",
  "kts",
  "scala",
  "sh",
  "bash",
  "zsh",
  "fish",
  "ps1",
  "bat",
  "cmd",
  "sql",
  "html",
  "htm",
  "css",
  "scss",
  "sass",
  "less",
  "vue",
  "svelte",
  "astro",
  "dart",
  "lua",
  // structured/configuration data
  "csv",
  "tsv",
  "json",
  "jsonl",
  "ndjson",
  "xml",
  "yaml",
  "yml",
  "toml",
  "ini",
  "cfg",
  "conf",
  "log",
  "properties",
  "env",
] as const;

const locallyReadableExtensions = new Set<string>([
  "txt",
  "md",
  "markdown",
  "rtf",
  "c",
  "cc",
  "cpp",
  "cxx",
  "h",
  "hpp",
  "cs",
  "go",
  "java",
  "js",
  "jsx",
  "ts",
  "tsx",
  "py",
  "rb",
  "rs",
  "php",
  "swift",
  "kt",
  "kts",
  "scala",
  "sh",
  "bash",
  "zsh",
  "fish",
  "ps1",
  "bat",
  "cmd",
  "sql",
  "html",
  "htm",
  "css",
  "scss",
  "sass",
  "less",
  "vue",
  "svelte",
  "astro",
  "dart",
  "lua",
  "csv",
  "tsv",
  "json",
  "jsonl",
  "ndjson",
  "xml",
  "yaml",
  "yml",
  "toml",
  "ini",
  "cfg",
  "conf",
  "log",
  "properties",
  "env",
]);

export const supportedFileAccept = supportedFileExtensions
  .map((extension) => `.${extension}`)
  .join(",");

export function getFileExtension(filename: string): string {
  const basename = filename.trim().toLowerCase().split(/[\\/]/).pop() || "";
  const index = basename.lastIndexOf(".");
  return index >= 0 ? basename.slice(index + 1) : "";
}

export function isSupportedFile(file: Pick<File, "name" | "type">): boolean {
  const extension = getFileExtension(file.name);
  return (
    supportedFileExtensions.includes(
      extension as (typeof supportedFileExtensions)[number],
    ) ||
    file.type.startsWith("image/") ||
    file.type.startsWith("audio/") ||
    file.type.startsWith("text/")
  );
}

export function isImageFile(file: Pick<File, "name" | "type">): boolean {
  return (
    file.type.startsWith("image/") ||
    getFileExtension(file.name) in imageMimeTypes
  );
}

export function isLocallyReadableFile(
  file: Pick<File, "name" | "type">,
): boolean {
  return (
    file.type.startsWith("text/") ||
    locallyReadableExtensions.has(getFileExtension(file.name))
  );
}

export async function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => {
      let result = reader.result as string;
      if (isImageFile(file) && result.startsWith("data:;base64,")) {
        const mimeType = imageMimeTypes[getFileExtension(file.name)];
        if (mimeType)
          result = result.replace("data:;base64,", `data:${mimeType};base64,`);
      }
      resolve(result);
    };
    reader.onerror = () => reject(new Error("Failed to read file"));
  });
}

export function checkFileSuffix(
  filename: string,
  suffixes: string | string[],
): boolean {
  filename = filename.toLowerCase();

  if (typeof suffixes === "string") {
    return filename.endsWith(suffixes);
  }

  return suffixes.some((suffix) => filename.endsWith(suffix));
}

export async function quickBlobParser(
  file: File,
  model: Model,
  onProgress?: (progress: number) => void,
): Promise<string> {
  // this function is used to parse the file quickly in local
  // otherwise, it will be parsed as a file

  if (file.size === 0 || file.name.length === 0) {
    throw new Error("File is empty");
  }

  const image = isImageFile(file);

  // Vision models can consume the original image directly. This path does not
  // depend on the document/OCR parser and also supports files without a MIME type.
  if (image && model.vision_model) {
    console.log("[parser] hit image file, using local vision parser");
    return fileToBase64(file);
  }

  if (image && !model.ocr_model && !model.reverse_model) {
    throw new Error(ImageModelRequiredError);
  }

  if (!model.reverse_model) {
    try {
      // Text, source code and structured data are directly readable by the
      // model and should not depend on the remote document parser.
      if (isLocallyReadableFile(file)) {
        console.log("[parser] hit locally readable file, using local parser");
        return await file.text();
      }
      console.log(file.type);
    } catch (e) {
      console.error(
        "[parser] local parser failed, switch to server parser: ",
        e,
      );
    }
  }

  return blobParser(file, model, onProgress);
}

export async function blobParser(
  file: File,
  model: Model,
  onProgress?: (progress: number) => void,
): Promise<string> {
  const endpoint = trimSuffixes(blobEndpoint, ["/upload", "/"]);

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const formData = new FormData();
    formData.append("file", file);
    formData.append("model", model.id);
    formData.append("enable_ocr", (model.ocr_model ?? false).toString());
    formData.append("enable_vision", (model.vision_model ?? false).toString());
    formData.append("save_all", (model.reverse_model ?? false).toString());
    xhr.open("POST", `${endpoint}/upload`, true);
    xhr.timeout = UploadTimeout;
    xhr.upload.onprogress = (progressEvent) => {
      console.debug(progressEvent);
      if (progressEvent.lengthComputable) {
        const percentCompleted = Math.round(
          (progressEvent.loaded * 100) / progressEvent.total,
        );
        console.debug(percentCompleted);
        onProgress?.(percentCompleted);
      }
    };
    xhr.onload = () => {
      let data: BlobParserResponse | undefined;
      try {
        data = JSON.parse(xhr.responseText) as BlobParserResponse;
      } catch {
        // Reverse proxies may return plain text or HTML when the request fails.
      }

      if (xhr.status >= 200 && xhr.status < 300) {
        if (!data) {
          reject(new Error("Invalid JSON response"));
        } else if (!data.status) {
          reject(
            new Error(
              data.error?.includes("requires a vision-capable model")
                ? ImageModelRequiredError
                : data.error || "The parser rejected this file",
            ),
          );
        } else if (
          typeof data.content !== "string" ||
          data.content.length === 0
        ) {
          reject(new Error("Result is empty"));
        } else {
          resolve(data.content);
        }
      } else {
        const reason =
          data?.error ||
          xhr.responseText.trim().slice(0, 300) ||
          xhr.statusText;
        reject(
          new Error(
            reason.includes("requires a vision-capable model")
              ? ImageModelRequiredError
              : `Document parser returned HTTP ${xhr.status}${
                  reason ? `: ${reason}` : ""
                }`,
          ),
        );
      }
    };
    xhr.onerror = () => {
      reject(new Error("Network error"));
    };
    xhr.ontimeout = () => {
      reject(new Error("Upload timed out"));
    };
    xhr.onabort = () => {
      reject(new Error("Upload was cancelled"));
    };
    xhr.send(formData);
  });
}
