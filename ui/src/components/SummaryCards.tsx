import React from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  LinearProgress,
  Chip,
} from '@mui/material';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';
import ErrorIcon from '@mui/icons-material/Error';
import StorageIcon from '@mui/icons-material/Storage';
import { Summary, FilterType } from '../types';

interface SummaryCardsProps {
  summary: Summary;
  scanTime: string;
  filter: FilterType;
  onFilterChange: (filter: FilterType) => void;
}

interface StatCardProps {
  title: string;
  value: number;
  total: number;
  icon: React.ReactNode;
  color: string;
  filterType: FilterType;
  currentFilter: FilterType;
  onClick: () => void;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  total,
  icon,
  color,
  filterType,
  currentFilter,
  onClick,
}) => {
  const percentage = total > 0 ? (value / total) * 100 : 0;
  const isActive = currentFilter === filterType;

  return (
    <Paper
      sx={{
        p: 2,
        cursor: 'pointer',
        border: isActive ? `2px solid ${color}` : '2px solid transparent',
        transition: 'all 0.2s',
        '&:hover': {
          transform: 'translateY(-2px)',
          boxShadow: 3,
        },
      }}
      onClick={onClick}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
        <Typography variant="body2" color="text.secondary">
          {title}
        </Typography>
        <Box sx={{ color }}>{icon}</Box>
      </Box>
      <Typography variant="h4" fontWeight="bold" sx={{ color }}>
        {value}
      </Typography>
      <LinearProgress
        variant="determinate"
        value={percentage}
        sx={{
          mt: 1,
          height: 6,
          borderRadius: 3,
          backgroundColor: 'grey.200',
          '& .MuiLinearProgress-bar': {
            backgroundColor: color,
            borderRadius: 3,
          },
        }}
      />
      <Typography variant="caption" color="text.secondary">
        {percentage.toFixed(1)}% of total
      </Typography>
    </Paper>
  );
};

export const SummaryCards: React.FC<SummaryCardsProps> = ({
  summary,
  scanTime,
  filter,
  onFilterChange,
}) => {
  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return dateString;
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h6">Scan Summary</Typography>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <Typography variant="body2" color="text.secondary">
            Last scan: {formatDate(scanTime)}
          </Typography>
          {filter !== 'all' && (
            <Chip
              label={`Filtered: ${filter}`}
              onDelete={() => onFilterChange('all')}
              size="small"
            />
          )}
        </Box>
      </Box>

      <Grid container spacing={2}>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Total Images"
            value={summary.total}
            total={summary.total}
            icon={<StorageIcon />}
            color="#1976d2"
            filterType="all"
            currentFilter={filter}
            onClick={() => onFilterChange('all')}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="ARM64 Compatible"
            value={summary.armCompatible}
            total={summary.total}
            icon={<CheckCircleIcon />}
            color="#2e7d32"
            filterType="compatible"
            currentFilter={filter}
            onClick={() => onFilterChange('compatible')}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Not Compatible"
            value={summary.notCompatible}
            total={summary.total}
            icon={<CancelIcon />}
            color="#d32f2f"
            filterType="incompatible"
            currentFilter={filter}
            onClick={() => onFilterChange('incompatible')}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Errors"
            value={summary.errors}
            total={summary.total}
            icon={<ErrorIcon />}
            color="#ed6c02"
            filterType="errors"
            currentFilter={filter}
            onClick={() => onFilterChange('errors')}
          />
        </Grid>
      </Grid>
    </Box>
  );
};
