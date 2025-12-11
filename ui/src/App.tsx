import React, { useState, useEffect, useCallback } from 'react';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import {
  Box,
  Typography,
  Button,
  CircularProgress,
  Alert,
  Paper,
  Stack,
  Divider,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import SearchIcon from '@mui/icons-material/Search';
import { ContextSelector } from './components/ContextSelector';
import { SummaryCards } from './components/SummaryCards';
import { ResultsTable } from './components/ResultsTable';
import { ExportButton } from './components/ExportButton';
import { InspectResponse, KubeContext, FilterType } from './types';

const ddClient = createDockerDesktopClient();

function App() {
  // State
  const [contexts, setContexts] = useState<KubeContext[]>([]);
  const [selectedContext, setSelectedContext] = useState<string>('');
  const [selectedNamespace, setSelectedNamespace] = useState<string>('all');
  const [inspectData, setInspectData] = useState<InspectResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadingContexts, setLoadingContexts] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterType>('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Fetch available Kubernetes contexts
  const fetchContexts = useCallback(async () => {
    setLoadingContexts(true);
    try {
      // Get contexts using kubectl config view (which outputs valid JSON)
      const result = await ddClient.extension.host?.cli.exec("kubectl", [
        "config",
        "view",
        "-o",
        "json"
      ]);

      if (!result || !result.stdout) {
        throw new Error('No output from kubectl');
      }

      // Parse kubectl JSON output
      const kubeconfigData = JSON.parse(result.stdout);
      
      // Extract contexts from kubeconfig
      const contextList: KubeContext[] = kubeconfigData.contexts?.map((ctx: any) => ({
        name: ctx.name,
        cluster: ctx.context?.cluster || '',
        user: ctx.context?.user || '',
        namespace: ctx.context?.namespace || 'default',
        current: ctx.name === kubeconfigData['current-context']
      })) || [];

      setContexts(contextList);

      // Set the current context as selected
      const currentContext = kubeconfigData['current-context'];
      if (currentContext) {
        setSelectedContext(currentContext);
      } else if (contextList.length > 0) {
        setSelectedContext(contextList[0].name);
      }

      setError(null);
    } catch (err: any) {
      console.error('Failed to fetch contexts:', err);
      const errorMessage = err?.stderr || err?.message || 'Failed to fetch Kubernetes contexts';
      setError(`${errorMessage}. Make sure kubectl is configured and a kubeconfig file exists.`);
      setContexts([]);
    } finally {
      setLoadingContexts(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchContexts();
  }, [fetchContexts]);

  // Inspect resources
  const inspectResources = async () => {
    if (!selectedContext) {
      setError('Please select a Kubernetes context');
      return;
    }

    setLoading(true);
    setError(null);
    setInspectData(null);

    try {
      const params = new URLSearchParams({
        context: selectedContext,
        namespace: selectedNamespace,
      });
      
      const result = await ddClient.extension.vm?.service?.get(`/inspect?${params}`);
      
      if (result && typeof result === 'object') {
        setInspectData(result as InspectResponse);
      } else {
        setError('Invalid response format from backend');
      }
    } catch (err: any) {
      console.error('Inspection failed:', err);
      setError(err?.message || 'Failed to inspect resources. Please check your cluster connection.');
    } finally {
      setLoading(false);
    }
  };

  // Filter results based on current filter and search
  const filteredResults = inspectData?.results?.filter((result) => {
    // Apply filter
    if (filter === 'compatible' && !result.isArmCompatible) return false;
    if (filter === 'incompatible' && (result.isArmCompatible || result.error)) return false;
    if (filter === 'errors' && !result.error) return false;

    // Apply search
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      return (
        result.image.toLowerCase().includes(query) ||
        result.resourceName.toLowerCase().includes(query) ||
        result.namespace.toLowerCase().includes(query) ||
        result.resourceType.toLowerCase().includes(query)
      );
    }

    return true;
  }) || [];

  return (
    <Box sx={{ p: 3, minHeight: '100vh' }}>
      {/* Header */}
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 3 }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <span role="img" aria-label="cpu">🔍</span>
            KubeArchInspect
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Check if your Kubernetes cluster images support ARM64 architecture
          </Typography>
        </Box>
        {inspectData && (
          <ExportButton data={inspectData} />
        )}
      </Stack>

      {/* Context Selection */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Stack direction="row" spacing={2} alignItems="center" flexWrap="wrap">
          <ContextSelector
            contexts={contexts}
            selectedContext={selectedContext}
            onContextChange={setSelectedContext}
            selectedNamespace={selectedNamespace}
            onNamespaceChange={setSelectedNamespace}
            loading={loadingContexts}
          />
          
          <Button
            variant="contained"
            color="primary"
            onClick={inspectResources}
            disabled={loading || !selectedContext}
            startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <SearchIcon />}
            sx={{ minWidth: 180, height: 40 }}
          >
            {loading ? 'Inspecting...' : 'Inspect Cluster'}
          </Button>

          <Button
            variant="outlined"
            onClick={fetchContexts}
            disabled={loadingContexts}
            startIcon={<RefreshIcon />}
            sx={{ height: 40 }}
          >
            Refresh
          </Button>
        </Stack>
      </Paper>

      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {/* Results Section */}
      {inspectData && (
        <>
          {/* Summary Cards */}
          <SummaryCards
            summary={inspectData.summary}
            scanTime={inspectData.scanTime}
            filter={filter}
            onFilterChange={setFilter}
          />

          <Divider sx={{ my: 3 }} />

          {/* Results Table */}
          <ResultsTable
            results={filteredResults}
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            loading={loading}
          />
        </>
      )}

      {/* Empty State */}
      {!inspectData && !loading && !error && (
        <Paper sx={{ p: 6, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary" sx={{ mb: 2 }}>
            <span role="img" aria-label="cluster">☸️</span> Ready to Inspect
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Select a Kubernetes context and click "Inspect Cluster" to check ARM64 compatibility
          </Typography>
        </Paper>
      )}

      {/* Loading State */}
      {loading && (
        <Paper sx={{ p: 6, textAlign: 'center' }}>
          <CircularProgress size={48} sx={{ mb: 2 }} />
          <Typography variant="h6" color="text.secondary">
            Inspecting cluster images...
          </Typography>
          <Typography variant="body2" color="text.secondary">
            This may take a moment depending on the number of images
          </Typography>
        </Paper>
      )}
    </Box>
  );
}

export default App;
