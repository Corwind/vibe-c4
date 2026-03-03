import { useState, useRef, useCallback } from "react";
import { useAnalyzeProject } from "../hooks/useProjects";

export function ProjectUpload() {
  const [gitUrl, setGitUrl] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const analyzeProject = useAnalyzeProject();

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) {
      setSelectedFile(file);
      setGitUrl("");
    }
  }, []);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setGitUrl("");
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedFile) {
      analyzeProject.mutate({ file: selectedFile });
    } else if (gitUrl.trim()) {
      analyzeProject.mutate({ gitUrl: gitUrl.trim() });
    }
  };

  const hasInput = selectedFile !== null || gitUrl.trim().length > 0;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label
          htmlFor="git-url"
          className="block text-sm font-medium text-text"
        >
          Git Repository URL
        </label>
        <input
          id="git-url"
          type="url"
          value={gitUrl}
          onChange={(e) => {
            setGitUrl(e.target.value);
            setSelectedFile(null);
          }}
          placeholder="https://github.com/user/repo"
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
          disabled={analyzeProject.isPending}
        />
      </div>

      <div className="text-center text-sm text-text-muted">or</div>

      <div
        role="button"
        tabIndex={0}
        aria-label="Drop zone for file upload"
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            fileInputRef.current?.click();
          }
        }}
        className={`cursor-pointer rounded-lg border-2 border-dashed p-8 text-center transition-colors ${
          isDragOver
            ? "border-primary bg-primary/5"
            : "border-gray-300 hover:border-primary"
        }`}
      >
        <input
          ref={fileInputRef}
          type="file"
          onChange={handleFileChange}
          className="hidden"
          accept=".zip,.tar.gz,.tgz"
          data-testid="file-input"
        />
        {selectedFile ? (
          <p className="text-sm text-text">
            Selected: <strong>{selectedFile.name}</strong>
          </p>
        ) : (
          <p className="text-sm text-text-muted">
            Drag & drop a project archive here, or click to browse
          </p>
        )}
      </div>

      {analyzeProject.isError && (
        <div role="alert" className="text-sm text-red-600">
          Failed to analyze project. Please try again.
        </div>
      )}

      {analyzeProject.isSuccess && (
        <div role="status" className="text-sm text-green-600">
          Project analysis started successfully.
        </div>
      )}

      <button
        type="submit"
        disabled={!hasInput || analyzeProject.isPending}
        className="w-full rounded-md bg-primary px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {analyzeProject.isPending ? "Analyzing..." : "Analyze Project"}
      </button>
    </form>
  );
}
