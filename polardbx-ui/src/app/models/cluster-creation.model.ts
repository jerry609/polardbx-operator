export interface NodeResources {
  cpu: string;
  memory: string;
}

export interface ClusterNodeConfig extends Record<string, unknown> {
  replicas: number;
  resources: NodeResources;
}

export interface ClusterTopologyConfig extends Record<string, unknown> {
  cn: ClusterNodeConfig;
  dn: ClusterNodeConfig;
  gms: ClusterNodeConfig;
  cdc?: ClusterNodeConfig;
}

export interface StorageConfig extends Record<string, unknown> {
  storageClassName?: string;
  size?: string;
  accessMode?: string;
  accessModes?: string[];
}

export interface NetworkConfig extends Record<string, unknown> {
  serviceType?: string;
  loadBalancerClass?: string;
}

export interface SecurityConfig extends Record<string, unknown> {
  enableTLS?: boolean;
  secretName?: string;
}

export interface AdvancedConfig extends Record<string, unknown> {
  enableMonitoring?: boolean;
  enableBackup?: boolean;
  enableLogCollection?: boolean;
  customLabels?: Record<string, string>;
  customAnnotations?: Record<string, string>;
}

export interface ClusterCreationConfig extends Record<string, unknown> {
  name?: string;
  namespace?: string;
  description?: string;
  version?: string;
  topology: ClusterTopologyConfig;
  storage: StorageConfig;
  network?: NetworkConfig;
  security?: SecurityConfig;
  advanced?: AdvancedConfig;
}

export type ClusterCreationStep = Record<string, unknown>;

export interface ClusterTemplate {
  name: string;
  label?: string;
  icon?: string;
  recommended?: boolean;
  description?: string;
  config: ClusterCreationConfig;
}

export const CLUSTER_TEMPLATES: readonly ClusterTemplate[] = [
  {
    name: 'minimal',
    label: '最小化',
    icon: 'bolt',
    recommended: true,
    description: '单副本开发测试',
    config: {
      topology: {
        cn: { replicas: 1, resources: { cpu: '500m', memory: '1Gi' } },
        dn: { replicas: 1, resources: { cpu: '500m', memory: '1Gi' } },
        gms: { replicas: 1, resources: { cpu: '500m', memory: '1Gi' } }
      },
      storage: { storageClassName: 'standard', size: '20Gi', accessMode: 'ReadWriteOnce' },
      network: { serviceType: 'ClusterIP' },
      security: { enableTLS: false }
    }
  }
];

export const CREATION_STEPS: ClusterCreationStep[] = [];

export interface ResourcePreset {
  cpu: string;
  memory: string;
}

export const RESOURCE_PRESETS: readonly ResourcePreset[] = [
  { cpu: '500m', memory: '1Gi' },
  { cpu: '1', memory: '2Gi' },
  { cpu: '2', memory: '4Gi' }
];

export interface StorageClassOption {
  value: string;
  label: string;
  description?: string;
}

export const STORAGE_CLASSES: readonly StorageClassOption[] = [
  { value: 'standard', label: 'standard', description: '默认存储类' }
];

export interface ServiceTypeOption {
  value: string;
  label: string;
  icon: string;
  description?: string;
}

export const SERVICE_TYPES: readonly ServiceTypeOption[] = [
  { value: 'ClusterIP', label: 'ClusterIP', icon: 'lan', description: '仅集群内访问' },
  { value: 'NodePort', label: 'NodePort', icon: 'upload', description: '通过节点端口访问' },
  { value: 'LoadBalancer', label: 'LoadBalancer', icon: 'cloud', description: '通过云负载均衡访问' }
];


