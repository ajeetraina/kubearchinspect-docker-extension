export interface Resource {
  name: string;
  namespace: string;
  kind: string;
  isArmCompatible: boolean;
  image?: string;
}