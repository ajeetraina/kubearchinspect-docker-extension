import React from 'react';
import {
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  SelectChangeEvent,
  CircularProgress,
  Box,
  Chip,
} from '@mui/material';
import { KubeContext } from '../types';

interface ContextSelectorProps {
  contexts: KubeContext[];
  selectedContext: string;
  onContextChange: (context: string) => void;
  selectedNamespace: string;
  onNamespaceChange: (namespace: string) => void;
  loading: boolean;
}

export const ContextSelector: React.FC<ContextSelectorProps> = ({
  contexts,
  selectedContext,
  onContextChange,
  selectedNamespace,
  onNamespaceChange,
  loading,
}) => {
  const handleContextChange = (event: SelectChangeEvent) => {
    onContextChange(event.target.value);
  };

  const handleNamespaceChange = (event: SelectChangeEvent) => {
    onNamespaceChange(event.target.value);
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <CircularProgress size={20} />
        <span>Loading contexts...</span>
      </Box>
    );
  }

  return (
    <>
      <FormControl sx={{ minWidth: 250 }} size="small">
        <InputLabel id="context-label">Kubernetes Context</InputLabel>
        <Select
          labelId="context-label"
          value={selectedContext}
          label="Kubernetes Context"
          onChange={handleContextChange}
        >
          {contexts.map((ctx) => (
            <MenuItem key={ctx.name} value={ctx.name}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                {ctx.name}
                {ctx.isCurrent && (
                  <Chip label="current" size="small" color="primary" variant="outlined" />
                )}
              </Box>
            </MenuItem>
          ))}
        </Select>
      </FormControl>

      <FormControl sx={{ minWidth: 180 }} size="small">
        <InputLabel id="namespace-label">Namespace</InputLabel>
        <Select
          labelId="namespace-label"
          value={selectedNamespace}
          label="Namespace"
          onChange={handleNamespaceChange}
        >
          <MenuItem value="all">All Namespaces</MenuItem>
          <MenuItem value="default">default</MenuItem>
          <MenuItem value="kube-system">kube-system</MenuItem>
          <MenuItem value="kube-public">kube-public</MenuItem>
        </Select>
      </FormControl>
    </>
  );
};
