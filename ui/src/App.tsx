import { useState, useEffect } from 'react';
import { createDockerDesktopClient } from '@docker/extension-api-client';
import { Box, Typography, Button, CircularProgress } from '@mui/material';
import { Statistics } from './components/Statistics';
import { FilterableResourceTable } from './components/FilterableResourceTable';

// Content of App.tsx as provided above...