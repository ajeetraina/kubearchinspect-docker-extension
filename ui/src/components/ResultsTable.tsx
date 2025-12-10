import React, { useState } from 'react';
import {
  Box,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  TextField,
  InputAdornment,
  Chip,
  Typography,
  Tooltip,
  IconButton,
  Collapse,
  TablePagination,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import { ImageResult } from '../types';

interface ResultsTableProps {
  results: ImageResult[];
  searchQuery: string;
  onSearchChange: (query: string) => void;
  loading: boolean;
}

type SortKey = 'image' | 'resourceName' | 'namespace' | 'resourceType' | 'isArmCompatible';
type SortDirection = 'asc' | 'desc';

export const ResultsTable: React.FC<ResultsTableProps> = ({
  results,
  searchQuery,
  onSearchChange,
  loading,
}) => {
  const [sortKey, setSortKey] = useState<SortKey>('image');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(25);

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDirection('asc');
    }
  };

  const sortedResults = [...results].sort((a, b) => {
    let aVal: any = a[sortKey];
    let bVal: any = b[sortKey];

    if (typeof aVal === 'boolean') {
      aVal = aVal ? 1 : 0;
      bVal = bVal ? 1 : 0;
    } else if (typeof aVal === 'string') {
      aVal = aVal.toLowerCase();
      bVal = bVal.toLowerCase();
    }

    if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
    if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
    return 0;
  });

  const paginatedResults = sortedResults.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  );

  const toggleRow = (image: string) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(image)) {
      newExpanded.delete(image);
    } else {
      newExpanded.add(image);
    }
    setExpandedRows(newExpanded);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getStatusChip = (result: ImageResult) => {
    if (result.error) {
      return (
        <Tooltip title={result.error}>
          <Chip label="Error" color="warning" size="small" />
        </Tooltip>
      );
    }
    if (result.isArmCompatible) {
      return <Chip label="ARM64 ✓" color="success" size="small" />;
    }
    return <Chip label="No ARM64" color="error" size="small" />;
  };

  const getArchChips = (architectures: string[]) => {
    if (!architectures || architectures.length === 0) {
      return <Typography variant="caption" color="text.secondary">Unknown</Typography>;
    }

    return (
      <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
        {architectures.map((arch) => (
          <Chip
            key={arch}
            label={arch}
            size="small"
            variant="outlined"
            color={arch.includes('arm') ? 'success' : 'default'}
            sx={{ fontSize: '0.7rem' }}
          />
        ))}
      </Box>
    );
  };

  return (
    <Paper sx={{ width: '100%', overflow: 'hidden' }}>
      <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6">
          Results ({results.length} images)
        </Typography>
        <TextField
          size="small"
          placeholder="Search images, resources..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon fontSize="small" />
              </InputAdornment>
            ),
          }}
          sx={{ width: 300 }}
        />
      </Box>

      <TableContainer sx={{ maxHeight: 600 }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox" />
              <TableCell>
                <TableSortLabel
                  active={sortKey === 'image'}
                  direction={sortKey === 'image' ? sortDirection : 'asc'}
                  onClick={() => handleSort('image')}
                >
                  Image
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortKey === 'resourceType'}
                  direction={sortKey === 'resourceType' ? sortDirection : 'asc'}
                  onClick={() => handleSort('resourceType')}
                >
                  Resource Type
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortKey === 'resourceName'}
                  direction={sortKey === 'resourceName' ? sortDirection : 'asc'}
                  onClick={() => handleSort('resourceName')}
                >
                  Resource Name
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortKey === 'namespace'}
                  direction={sortKey === 'namespace' ? sortDirection : 'asc'}
                  onClick={() => handleSort('namespace')}
                >
                  Namespace
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortKey === 'isArmCompatible'}
                  direction={sortKey === 'isArmCompatible' ? sortDirection : 'asc'}
                  onClick={() => handleSort('isArmCompatible')}
                >
                  Status
                </TableSortLabel>
              </TableCell>
              <TableCell>Architectures</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {paginatedResults.map((result, index) => (
              <React.Fragment key={`${result.image}-${index}`}>
                <TableRow
                  hover
                  sx={{
                    '&:last-child td, &:last-child th': { border: 0 },
                    backgroundColor: result.error
                      ? 'warning.light'
                      : result.isArmCompatible
                      ? 'inherit'
                      : 'error.light',
                    opacity: result.error ? 0.8 : 1,
                  }}
                >
                  <TableCell padding="checkbox">
                    <IconButton size="small" onClick={() => toggleRow(result.image)}>
                      {expandedRows.has(result.image) ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                    </IconButton>
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Tooltip title={result.image}>
                        <Typography
                          variant="body2"
                          sx={{
                            maxWidth: 300,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          }}
                        >
                          {result.image}
                        </Typography>
                      </Tooltip>
                      <Tooltip title="Copy image name">
                        <IconButton size="small" onClick={() => copyToClipboard(result.image)}>
                          <ContentCopyIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip label={result.resourceType} size="small" variant="outlined" />
                  </TableCell>
                  <TableCell>{result.resourceName}</TableCell>
                  <TableCell>
                    <Chip label={result.namespace} size="small" color="info" variant="outlined" />
                  </TableCell>
                  <TableCell>{getStatusChip(result)}</TableCell>
                  <TableCell>{getArchChips(result.supportedArch)}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={7}>
                    <Collapse in={expandedRows.has(result.image)} timeout="auto" unmountOnExit>
                      <Box sx={{ py: 2, px: 3, backgroundColor: 'grey.50' }}>
                        <Typography variant="subtitle2" gutterBottom>
                          Full Image Path:
                        </Typography>
                        <Typography
                          variant="body2"
                          sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}
                        >
                          {result.image}
                        </Typography>
                        {result.error && (
                          <>
                            <Typography variant="subtitle2" sx={{ mt: 1 }} gutterBottom>
                              Error:
                            </Typography>
                            <Typography variant="body2" color="error">
                              {result.error}
                            </Typography>
                          </>
                        )}
                        {result.supportedArch && result.supportedArch.length > 0 && (
                          <>
                            <Typography variant="subtitle2" sx={{ mt: 1 }} gutterBottom>
                              Supported Architectures:
                            </Typography>
                            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                              {result.supportedArch.map((arch) => (
                                <Chip
                                  key={arch}
                                  label={arch}
                                  color={arch.includes('arm') ? 'success' : 'default'}
                                  size="small"
                                />
                              ))}
                            </Box>
                          </>
                        )}
                      </Box>
                    </Collapse>
                  </TableCell>
                </TableRow>
              </React.Fragment>
            ))}
            {results.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                  <Typography color="text.secondary">
                    No results found
                  </Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <TablePagination
        rowsPerPageOptions={[10, 25, 50, 100]}
        component="div"
        count={results.length}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={(_, newPage) => setPage(newPage)}
        onRowsPerPageChange={(e) => {
          setRowsPerPage(parseInt(e.target.value, 10));
          setPage(0);
        }}
      />
    </Paper>
  );
};
