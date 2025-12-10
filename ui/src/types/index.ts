// Types matching the backend API response

export interface ImageResult {
  image: string;
  isArmCompatible: boolean;
  supportedArch: string[];
  resourceType: string;
  resourceName: string;
  namespace: string;
  error?: string;
}

export interface Summary {
  total: number;
  armCompatible: number;
  notCompatible: number;
  errors: number;
}

export interface InspectResponse {
  results: ImageResult[];
  summary: Summary;
  scanTime: string;
  context: string;
  namespace: string;
}

export interface KubeContext {
  name: string;
  cluster: string;
  isCurrent: boolean;
}

export interface ContextsResponse {
  contexts: KubeContext[];
  current: string;
}

export type FilterType = 'all' | 'compatible' | 'incompatible' | 'errors';
