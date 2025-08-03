/**
 * Admin Reports Management JavaScript
 */

// Global admin reports manager
window.AdminReportsManager = {
    charts: {},
    
    // Initialize admin reports functionality
    init: function() {
        this.initEventListeners();
        this.initCharts();
        this.initDateRangePicker();
    },
    
    // Initialize event listeners
    initEventListeners: function() {
        // Export all reports
        window.exportAllReports = () => {
            this.exportAllReports();
        };
        
        // Generate reports
        window.generateReports = () => {
            this.generateReports();
        };
        
        // Individual report functions
        window.generateSalesReport = () => {
            this.generateSalesReport();
        };
        
        window.generateProductsReport = () => {
            this.generateProductsReport();
        };
        
        window.generateCustomerReport = () => {
            this.generateCustomerReport();
        };
        
        window.generateVendorReport = () => {
            this.generateVendorReport();
        };
        
        window.refreshDetailedReports = () => {
            this.refreshDetailedReports();
        };
        
        // Report actions
        window.viewReport = (reportId) => {
            this.viewReport(reportId);
        };
        
        window.downloadReport = (reportId) => {
            this.downloadReport(reportId);
        };
        
        window.scheduleReport = (reportId) => {
            this.scheduleReport(reportId);
        };
        
        window.deleteReport = (reportId) => {
            this.deleteReport(reportId);
        };
        
        // Schedule modal functions
        window.showScheduleModal = () => {
            this.showScheduleModal();
        };
        
        window.closeScheduleModal = () => {
            this.closeScheduleModal();
        };
        
        window.saveSchedule = () => {
            this.saveSchedule();
        };
        
        // Schedule management
        window.editSchedule = (scheduleId) => {
            this.editSchedule(scheduleId);
        };
        
        window.toggleSchedule = (scheduleId) => {
            this.toggleSchedule(scheduleId);
        };
        
        window.deleteSchedule = (scheduleId) => {
            this.deleteSchedule(scheduleId);
        };
    },
    
    // Initialize charts
    initCharts: function() {
        this.initSalesChart();
        this.initCustomerChart();
    },
    
    // Initialize sales chart
    initSalesChart: function() {
        const ctx = document.getElementById('salesChart');
        if (!ctx) return;
        
        this.charts.sales = new Chart(ctx, {
            type: 'line',
            data: {
                labels: ['Pzt', 'Sal', 'Çar', 'Per', 'Cum', 'Cmt', 'Paz'],
                datasets: [{
                    label: 'Satışlar',
                    data: [12000, 19000, 15000, 25000, 22000, 30000, 28000],
                    borderColor: 'rgb(59, 130, 246)',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return '₺' + value.toLocaleString();
                            }
                        }
                    }
                }
            }
        });
    },
    
    // Initialize customer chart
    initCustomerChart: function() {
        const ctx = document.getElementById('customerChart');
        if (!ctx) return;
        
        this.charts.customer = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Yeni Müşteriler', 'Geri Dönen Müşteriler'],
                datasets: [{
                    data: [123, 456],
                    backgroundColor: [
                        'rgb(59, 130, 246)',
                        'rgb(16, 185, 129)'
                    ]
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
    },
    
    // Initialize date range picker
    initDateRangePicker: function() {
        const dateRange = document.getElementById('dateRange');
        if (dateRange) {
            dateRange.addEventListener('change', () => {
                this.updateReportsForDateRange();
            });
        }
    },
    
    // Export all reports
    exportAllReports: function() {
        const format = document.getElementById('exportFormat')?.value || 'pdf';
        
        this.showLoading('Tüm raporlar dışa aktarılıyor...');
        
        fetch('/api/admin/reports/export-all', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                format: format,
                date_range: document.getElementById('dateRange')?.value
            })
        })
        .then(response => response.blob())
        .then(blob => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `all-reports-${new Date().toISOString().split('T')[0]}.${format}`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            this.hideLoading();
            this.showAlert('Raporlar başarıyla dışa aktarıldı', 'success');
        })
        .catch(error => {
            this.hideLoading();
            this.showAlert('Raporlar dışa aktarılırken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Generate reports
    generateReports: function() {
        this.showLoading('Raporlar güncelleniyor...');
        
        const dateRange = document.getElementById('dateRange')?.value;
        const reportType = document.getElementById('reportType')?.value;
        
        fetch('/api/admin/reports/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                date_range: dateRange,
                report_type: reportType
            })
        })
        .then(response => response.json())
        .then(data => {
            this.hideLoading();
            if (data.success) {
                this.showAlert('Raporlar başarıyla güncellendi', 'success');
                this.refreshReports();
            } else {
                this.showAlert('Raporlar güncellenirken hata oluştu', 'error');
            }
        })
        .catch(error => {
            this.hideLoading();
            this.showAlert('Raporlar güncellenirken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Generate individual reports
    generateSalesReport: function() {
        this.generateIndividualReport('sales', 'Satış raporu');
    },
    
    generateProductsReport: function() {
        this.generateIndividualReport('products', 'Ürün raporu');
    },
    
    generateCustomerReport: function() {
        this.generateIndividualReport('customers', 'Müşteri raporu');
    },
    
    generateVendorReport: function() {
        this.generateIndividualReport('vendors', 'Satıcı raporu');
    },
    
    // Generate individual report
    generateIndividualReport: function(type, name) {
        this.showLoading(`${name} güncelleniyor...`);
        
        fetch(`/api/admin/reports/generate/${type}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        })
        .then(response => response.json())
        .then(data => {
            this.hideLoading();
            if (data.success) {
                this.showAlert(`${name} başarıyla güncellendi`, 'success');
                this.refreshReportSection(type);
            } else {
                this.showAlert(`${name} güncellenirken hata oluştu`, 'error');
            }
        })
        .catch(error => {
            this.hideLoading();
            this.showAlert(`${name} güncellenirken hata oluştu: ` + error.message, 'error');
        });
    },
    
    // Export individual report
    exportReport: function(type) {
        const format = document.getElementById('exportFormat')?.value || 'pdf';
        
        fetch(`/api/admin/reports/export/${type}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ format: format })
        })
        .then(response => response.blob())
        .then(blob => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${type}-report-${new Date().toISOString().split('T')[0]}.${format}`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            this.showAlert('Rapor başarıyla dışa aktarıldı', 'success');
        })
        .catch(error => {
            this.showAlert('Rapor dışa aktarılırken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Refresh detailed reports
    refreshDetailedReports: function() {
        this.showLoading('Detaylı raporlar yenileniyor...');
        
        fetch('/api/admin/reports/detailed')
        .then(response => response.json())
        .then(data => {
            this.hideLoading();
            if (data.success) {
                this.updateDetailedReportsTable(data.reports);
            }
        })
        .catch(error => {
            this.hideLoading();
            this.showAlert('Raporlar yenilenirken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // View report
    viewReport: function(reportId) {
        window.open(`/admin/reports/view/${reportId}`, '_blank');
    },
    
    // Download report
    downloadReport: function(reportId) {
        fetch(`/api/admin/reports/download/${reportId}`)
        .then(response => response.blob())
        .then(blob => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `report-${reportId}.pdf`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        })
        .catch(error => {
            this.showAlert('Rapor indirilirken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Schedule report
    scheduleReport: function(reportId) {
        // Implementation for scheduling existing report
        this.showAlert('Rapor zamanlama özelliği yakında eklenecek', 'info');
    },
    
    // Delete report
    deleteReport: function(reportId) {
        if (!confirm('Bu raporu silmek istediğinizden emin misiniz?')) {
            return;
        }
        
        fetch(`/api/admin/reports/${reportId}`, {
            method: 'DELETE'
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.showAlert('Rapor başarıyla silindi', 'success');
                this.refreshDetailedReports();
            } else {
                this.showAlert('Rapor silinirken hata oluştu', 'error');
            }
        })
        .catch(error => {
            this.showAlert('Rapor silinirken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Schedule modal functions
    showScheduleModal: function() {
        document.getElementById('scheduleModal').classList.remove('hidden');
    },
    
    closeScheduleModal: function() {
        document.getElementById('scheduleModal').classList.add('hidden');
        document.getElementById('scheduleForm').reset();
    },
    
    saveSchedule: function() {
        const form = document.getElementById('scheduleForm');
        const formData = new FormData(form);
        
        const scheduleData = {
            report_type: formData.get('report_type'),
            schedule: formData.get('schedule'),
            recipients: formData.get('recipients'),
            format: formData.get('format')
        };
        
        if (!scheduleData.report_type || !scheduleData.schedule || !scheduleData.recipients) {
            this.showAlert('Lütfen tüm gerekli alanları doldurun', 'error');
            return;
        }
        
        fetch('/api/admin/reports/schedule', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(scheduleData)
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.showAlert('Rapor zamanlaması başarıyla oluşturuldu', 'success');
                this.closeScheduleModal();
                this.refreshScheduledReports();
            } else {
                this.showAlert('Rapor zamanlaması oluşturulurken hata oluştu', 'error');
            }
        })
        .catch(error => {
            this.showAlert('Rapor zamanlaması oluşturulurken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Schedule management functions
    editSchedule: function(scheduleId) {
        // Implementation for editing schedule
        this.showAlert('Zamanlama düzenleme özelliği yakında eklenecek', 'info');
    },
    
    toggleSchedule: function(scheduleId) {
        fetch(`/api/admin/reports/schedule/${scheduleId}/toggle`, {
            method: 'POST'
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.showAlert('Zamanlama durumu güncellendi', 'success');
                this.refreshScheduledReports();
            } else {
                this.showAlert('Zamanlama durumu güncellenirken hata oluştu', 'error');
            }
        })
        .catch(error => {
            this.showAlert('Zamanlama durumu güncellenirken hata oluştu: ' + error.message, 'error');
        });
    },
    
    deleteSchedule: function(scheduleId) {
        if (!confirm('Bu zamanlamayı silmek istediğinizden emin misiniz?')) {
            return;
        }
        
        fetch(`/api/admin/reports/schedule/${scheduleId}`, {
            method: 'DELETE'
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                this.showAlert('Zamanlama başarıyla silindi', 'success');
                this.refreshScheduledReports();
            } else {
                this.showAlert('Zamanlama silinirken hata oluştu', 'error');
            }
        })
        .catch(error => {
            this.showAlert('Zamanlama silinirken hata oluştu: ' + error.message, 'error');
        });
    },
    
    // Helper functions
    updateReportsForDateRange: function() {
        // Update charts and data based on selected date range
        this.refreshReports();
    },
    
    refreshReports: function() {
        // Refresh all report sections
        this.refreshReportSection('sales');
        this.refreshReportSection('products');
        this.refreshReportSection('customers');
        this.refreshReportSection('vendors');
        this.refreshDetailedReports();
    },
    
    refreshReportSection: function(type) {
        // Refresh specific report section
        // Implementation would update the specific section
    },
    
    refreshScheduledReports: function() {
        // Refresh scheduled reports list
        // Implementation would update the scheduled reports section
    },
    
    updateDetailedReportsTable: function(reports) {
        // Update detailed reports table
        // Implementation would update the table with new data
    },
    
    showLoading: function(message) {
        // Show loading indicator
        console.log('Loading:', message);
    },
    
    hideLoading: function() {
        // Hide loading indicator
        console.log('Loading complete');
    },
    
    showAlert: function(message, type) {
        // Show alert message
        alert(message);
    }
};

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    window.AdminReportsManager.init();
});