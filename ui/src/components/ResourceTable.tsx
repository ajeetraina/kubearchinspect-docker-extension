import React from 'react';
import { Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';

export const ResourceTable: React.FC<{ resources: Resource[] }> = ({ resources }) => {
    return (
        <Table>
            <TableHead>
                <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Kind</TableCell>
                    <TableCell>Namespace</TableCell>
                    <TableCell>ARM Compatible</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {resources.map(resource => (
                    <TableRow key={`${resource.namespace}-${resource.name}`}>
                        <TableCell>{resource.name}</TableCell>
                        <TableCell>{resource.kind}</TableCell>
                        <TableCell>{resource.namespace}</TableCell>
                        <TableCell>
                            {resource.isArmCompatible ? '✅' : '❌'}
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    );
};