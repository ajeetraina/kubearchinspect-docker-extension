import React from 'react';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import { Box, Typography } from '@mui/material';
import { Statistics } from './components/Statistics';
import { FilterableResourceTable } from './components/FilterableResourceTable';

// Implementation as provided above