// ===========================================
// SYSTÈME DE NOTIFICATIONS PERSONNALISÉ
// ===========================================

class NotificationSystem {
    constructor() {
        this.container = null;
        this.init();
    }

    init() {
        // Créer le conteneur de notifications s'il n'existe pas
        this.container = document.getElementById('notification-container');
        if (!this.container) {
            this.container = document.createElement('div');
            this.container.id = 'notification-container';
            this.container.className = 'notification-container';
            document.body.appendChild(this.container);
        }
    }

    show(message, type = 'info', title = null, duration = 5000) {
        const notification = this.createNotification(message, type, title);
        this.container.appendChild(notification);

        // Animation d'entrée
        setTimeout(() => {
            notification.classList.add('show');
        }, 10);

        // Auto-suppression
        if (duration > 0) {
            setTimeout(() => {
                this.remove(notification);
            }, duration);
        }

        return notification;
    }

    createNotification(message, type, title) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;

        const icons = {
            success: 'fas fa-check-circle',
            error: 'fas fa-exclamation-circle',
            warning: 'fas fa-exclamation-triangle',
            info: 'fas fa-info-circle'
        };

        const titles = {
            success: title || 'Succès',
            error: title || 'Erreur',
            warning: title || 'Attention',
            info: title || 'Information'
        };

        notification.innerHTML = `
            <div class="notification-icon">
                <i class="${icons[type] || icons.info}"></i>
            </div>
            <div class="notification-content">
                <div class="notification-title">${titles[type]}</div>
                <div class="notification-message">${message}</div>
            </div>
            <button class="notification-close" onclick="notifications.remove(this.parentElement)">
                <i class="fas fa-times"></i>
            </button>
        `;

        return notification;
    }

    remove(notification) {
        if (notification && notification.parentElement) {
            notification.classList.remove('show');
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.parentElement.removeChild(notification);
                }
            }, 300);
        }
    }

    success(message, title = null, duration = 5000) {
        return this.show(message, 'success', title, duration);
    }

    error(message, title = null, duration = 8000) {
        return this.show(message, 'error', title, duration);
    }

    warning(message, title = null, duration = 6000) {
        return this.show(message, 'warning', title, duration);
    }

    info(message, title = null, duration = 5000) {
        return this.show(message, 'info', title, duration);
    }
}

// ===========================================
// SYSTÈME DE MODALES PERSONNALISÉ
// ===========================================

class ModalSystem {
    constructor() {
        this.currentModal = null;
    }

    confirm(message, title = 'Confirmation', options = {}) {
        return new Promise((resolve) => {
            const modal = this.createModal('confirm', title, message, [
                {
                    text: options.cancelText || 'Annuler',
                    class: 'btn btn-secondary',
                    action: () => {
                        this.close();
                        resolve(false);
                    }
                },
                {
                    text: options.confirmText || 'Confirmer',
                    class: 'btn btn-primary',
                    action: () => {
                        this.close();
                        resolve(true);
                    }
                }
            ]);
            this.show(modal);
        });
    }

    alert(message, title = 'Information', type = 'info') {
        return new Promise((resolve) => {
            const modal = this.createModal(type, title, message, [
                {
                    text: 'OK',
                    class: 'btn btn-primary',
                    action: () => {
                        this.close();
                        resolve();
                    }
                }
            ]);
            this.show(modal);
        });
    }

    prompt(message, title = 'Saisie', defaultValue = '', options = {}) {
        return new Promise((resolve) => {
            const inputId = 'modal-input-' + Date.now();
            const messageWithInput = `
                ${message}
                <div style="margin-top: 1rem;">
                    <input type="text" id="${inputId}" class="form-control" 
                           value="${defaultValue}" placeholder="${options.placeholder || ''}"
                           style="width: 100%; padding: 0.75rem; border: 1px solid #e5e7eb; border-radius: 8px;">
                </div>
            `;

            const modal = this.createModal('info', title, messageWithInput, [
                {
                    text: options.cancelText || 'Annuler',
                    class: 'btn btn-secondary',
                    action: () => {
                        this.close();
                        resolve(null);
                    }
                },
                {
                    text: options.confirmText || 'OK',
                    class: 'btn btn-primary',
                    action: () => {
                        const input = document.getElementById(inputId);
                        const value = input ? input.value : '';
                        this.close();
                        resolve(value);
                    }
                }
            ]);

            this.show(modal);

            // Focus sur l'input après affichage
            setTimeout(() => {
                const input = document.getElementById(inputId);
                if (input) {
                    input.focus();
                    input.select();
                }
            }, 100);
        });
    }

    createModal(type, title, message, buttons) {
        const overlay = document.createElement('div');
        overlay.className = 'modal-overlay';

        const icons = {
            confirm: 'fas fa-question-circle',
            error: 'fas fa-exclamation-circle',
            warning: 'fas fa-exclamation-triangle',
            info: 'fas fa-info-circle',
            success: 'fas fa-check-circle'
        };

        const buttonsHtml = buttons.map(button => 
            `<button class="${button.class}" data-action="${buttons.indexOf(button)}">${button.text}</button>`
        ).join('');

        overlay.innerHTML = `
            <div class="modal ${type}">
                <div class="modal-header">
                    <div class="modal-icon">
                        <i class="${icons[type] || icons.info}"></i>
                    </div>
                    <h3 class="modal-title">${title}</h3>
                </div>
                <div class="modal-body">
                    <div class="modal-message">${message}</div>
                </div>
                <div class="modal-footer">
                    ${buttonsHtml}
                </div>
            </div>
        `;

        // Ajouter les événements aux boutons
        overlay.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-overlay')) {
                // Clic en dehors de la modale
                if (buttons.length > 1) {
                    buttons[0].action(); // Action du premier bouton (généralement annuler)
                }
            } else if (e.target.hasAttribute('data-action')) {
                const actionIndex = parseInt(e.target.getAttribute('data-action'));
                if (buttons[actionIndex]) {
                    buttons[actionIndex].action();
                }
            }
        });

        // Gestion des touches clavier
        const keyHandler = (e) => {
            if (e.key === 'Escape' && buttons.length > 1) {
                buttons[0].action(); // Échap = annuler
            } else if (e.key === 'Enter' && buttons.length > 0) {
                buttons[buttons.length - 1].action(); // Entrée = confirmer
            }
        };

        overlay.addEventListener('keydown', keyHandler);
        document.addEventListener('keydown', keyHandler);

        // Nettoyer l'événement quand la modale est fermée
        overlay._cleanup = () => {
            document.removeEventListener('keydown', keyHandler);
        };

        return overlay;
    }

    show(modal) {
        if (this.currentModal) {
            this.close();
        }

        this.currentModal = modal;
        document.body.appendChild(modal);

        // Animation d'entrée
        setTimeout(() => {
            modal.classList.add('show');
        }, 10);
    }

    close() {
        if (this.currentModal) {
            this.currentModal.classList.remove('show');
            
            setTimeout(() => {
                if (this.currentModal && this.currentModal.parentElement) {
                    if (this.currentModal._cleanup) {
                        this.currentModal._cleanup();
                    }
                    this.currentModal.parentElement.removeChild(this.currentModal);
                }
                this.currentModal = null;
            }, 300);
        }
    }
}

// ===========================================
// INITIALISATION GLOBALE
// ===========================================

// Créer les instances globales
const notifications = new NotificationSystem();
const modals = new ModalSystem();

// Fonctions globales pour compatibilité
window.showNotification = (message, type = 'info', title = null) => {
    return notifications.show(message, type, title);
};

window.showSuccess = (message, title = null) => {
    return notifications.success(message, title);
};

window.showError = (message, title = null) => {
    return notifications.error(message, title);
};

window.showWarning = (message, title = null) => {
    return notifications.warning(message, title);
};

window.showInfo = (message, title = null) => {
    return notifications.info(message, title);
};

window.confirmAction = (message, title = 'Confirmation', options = {}) => {
    return modals.confirm(message, title, options);
};

window.alertMessage = (message, title = 'Information', type = 'info') => {
    return modals.alert(message, title, type);
};

window.promptUser = (message, title = 'Saisie', defaultValue = '', options = {}) => {
    return modals.prompt(message, title, defaultValue, options);
};

// Export pour utilisation en modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { notifications, modals };
} 