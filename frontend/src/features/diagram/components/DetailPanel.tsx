import type { C4NodeData } from "@/features/project/types/project.types";

const typeLabels: Record<string, string> = {
  system: "System",
  container: "Container",
  component: "Component",
  code: "Code",
  external: "External System",
  person: "Person",
};

const typeBorderColors: Record<string, string> = {
  system: "border-l-blue-600",
  container: "border-l-blue-400",
  component: "border-l-blue-200",
  code: "border-l-gray-400",
  external: "border-l-gray-400",
  person: "border-l-green-700",
};

interface DetailPanelProps {
  nodeData: C4NodeData | null;
  nodeType: string | null;
  onClose: () => void;
}

export function DetailPanel({ nodeData, nodeType, onClose }: DetailPanelProps) {
  if (!nodeData || !nodeType) {
    return null;
  }

  const typeLabel = typeLabels[nodeType] ?? nodeType;
  const borderColor = typeBorderColors[nodeType] ?? "border-l-gray-300";

  return (
    <aside className={`h-full w-80 overflow-y-auto border-l-4 bg-white shadow-lg ${borderColor}`}>
      <div className="flex items-center justify-between border-b border-gray-200 px-4 py-3">
        <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">
          {typeLabel}
        </span>
        <button
          onClick={onClose}
          aria-label="Close detail panel"
          className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
        >
          <svg
            className="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div className="space-y-4 px-4 py-4">
        <h2 className="text-lg font-bold text-text">
          {nodeData.label}
          {nodeData.isEntrypoint && (
            <span className="ml-2 inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800">
              {nodeData.entrypointKind === "http_handler" ? "HTTP Handler" : nodeData.entrypointKind === "main" ? "Main" : nodeData.entrypointKind || "Entrypoint"}
            </span>
          )}
        </h2>

        {nodeData.description && (
          <p className="text-sm leading-relaxed text-text-muted">
            {nodeData.description}
          </p>
        )}

        {nodeData.entrypointRoute && (
          <DetailField label="Route" value={nodeData.entrypointRoute} />
        )}

        {nodeData.systemKind && (
          <DetailField label="System Kind" value={nodeData.systemKind} />
        )}

        {nodeData.technology && (
          <DetailField label="Technology" value={nodeData.technology} />
        )}

        {nodeData.packagePath && (
          <DetailField label="Package" value={nodeData.packagePath} />
        )}

        {nodeData.role && <DetailField label="Role" value={nodeData.role} />}

        {nodeData.kind && <DetailField label="Kind" value={nodeData.kind} />}

        {nodeData.signature && (
          <DetailSection title="Signature">
            <code className="block whitespace-pre-wrap rounded bg-gray-50 px-3 py-2 font-mono text-xs text-gray-800">
              {nodeData.signature}
            </code>
          </DetailSection>
        )}

        {nodeData.methods && nodeData.methods.length > 0 && (
          <DetailSection title="Methods">
            <ul className="space-y-1">
              {nodeData.methods.map((method) => (
                <li
                  key={method}
                  className="rounded bg-gray-50 px-3 py-1.5 font-mono text-xs text-gray-700"
                >
                  {method}
                </li>
              ))}
            </ul>
          </DetailSection>
        )}

        {nodeData.fields && nodeData.fields.length > 0 && (
          <DetailSection title="Fields">
            <ul className="space-y-1">
              {nodeData.fields.map((field) => (
                <li
                  key={field}
                  className="rounded bg-gray-50 px-3 py-1.5 font-mono text-xs text-gray-700"
                >
                  {field}
                </li>
              ))}
            </ul>
          </DetailSection>
        )}

        {nodeData.parameters && nodeData.parameters.length > 0 && (
          <DetailSection title="Parameters">
            <ul className="space-y-1">
              {nodeData.parameters.map((param) => (
                <li
                  key={param}
                  className="rounded bg-gray-50 px-3 py-1.5 font-mono text-xs text-gray-700"
                >
                  {param}
                </li>
              ))}
            </ul>
          </DetailSection>
        )}

        {nodeData.returnTypes && nodeData.returnTypes.length > 0 && (
          <DetailSection title="Returns">
            <ul className="space-y-1">
              {nodeData.returnTypes.map((rt) => (
                <li
                  key={rt}
                  className="rounded bg-gray-50 px-3 py-1.5 font-mono text-xs text-gray-700"
                >
                  {rt}
                </li>
              ))}
            </ul>
          </DetailSection>
        )}

        {nodeData.dependencies && nodeData.dependencies.length > 0 && (
          <DetailSection title="Dependencies">
            <ul className="space-y-1">
              {nodeData.dependencies.map((dep) => (
                <li
                  key={dep}
                  className="rounded bg-gray-50 px-3 py-1.5 text-xs text-gray-700"
                >
                  {dep}
                </li>
              ))}
            </ul>
          </DetailSection>
        )}
      </div>
    </aside>
  );
}

function DetailField({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wider text-gray-400">
        {label}
      </dt>
      <dd className="mt-0.5 text-sm text-text">{value}</dd>
    </div>
  );
}

function DetailSection({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <h3 className="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400">
        {title}
      </h3>
      {children}
    </div>
  );
}
