import React from 'react';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableRow, 
  Paper, 
  TableContainer 
} from '@mui/material';
import { Resource } from '../types';

interface ResourceTableProps {
  resources: Resource[];
}

export const ResourceTable: React.FC<ResourceTableProps> = ({ resources }) => {
  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>Namespace</TableCell>
            <TableCell>Kind</TableCell>
            <TableCell>ARM Compatible</TableCell>
            <TableCell>Image</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {resources.map((resource, index) => (
            <TableRow key={`${resource.namespace}-${resource.name}-${index}`}>
              <TableCell>{resource.name}</TableCell>
              <TableCell>{resource.namespace}</TableCell>
              <TableCell>{resource.kind}</TableCell>
              <TableCell>
                {resource.isArmCompatible ? '✅' : '❌'}
              </TableCell>
              <TableCell>{resource.image || '-'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};