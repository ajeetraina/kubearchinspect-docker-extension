import React, { useState } from 'react';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import { Box, Typography, Button, CircularProgress } from '@mui/material';
import { ResourceTable } from './components/ResourceTable';
import { Resource } from './types';

const client = createDockerDesktopClient();

function App() {
  const [resources, setResources] = useState<Resource[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const inspectResources = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await client.extension.vm?.service?.get('/inspect');
      setResources(result);
    } catch (err) {
      console.error(err);
      setError('Failed to inspect resources. Please check your connection to the cluster.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ padding: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Kubernetes ARM Inspector
      </Typography>

      <Button 
        variant="contained" 
        onClick={inspectResources}
        disabled={loading}
        sx={{ mb: 3 }}
      >
        {loading ? <CircularProgress size={24} /> : 'Inspect Resources'}
      </Button>

      {error && (
        <Typography color="error" sx={{ mb: 2 }}>
          {error}
        </Typography>
      )}

      {resources.length > 0 && (
        <ResourceTable resources={resources} />
      )}
    </Box>
  );
}

export default App;