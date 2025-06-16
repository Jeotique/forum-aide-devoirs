// Forum d'aide aux devoirs - Script JavaScript

document.addEventListener('DOMContentLoaded', function() {
    console.log('🎓 Forum d\'aide aux devoirs chargé !');
    
    // Initialiser les fonctionnalités
    initVoting();
    initModeration();
    initFormValidation();
    initNotifications();
});

// === SYSTÈME DE VOTES ===
function initVoting() {
    // Gestion des likes/dislikes sur les posts et commentaires
    document.querySelectorAll('.vote-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            
            const voteType = this.dataset.type; // 'like' ou 'dislike'
            const targetType = this.dataset.target; // 'post' ou 'comment'
            const targetId = this.dataset.id;
            
            // Vérifier si l'utilisateur est connecté
            if (!document.body.classList.contains('logged-in')) {
                showNotification('Connectez-vous pour voter', 'warning');
                return;
            }
            
            // Envoyer le vote
            fetch('/api/vote', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams({
                    type: voteType,
                    target: targetType,
                    target_id: targetId
                })
            })
            .then(response => {
                if (response.ok) {
                    // Mettre à jour l'interface
                    updateVoteDisplay(targetType, targetId, voteType);
                    showNotification('Vote enregistré !', 'success');
                } else {
                    showNotification('Erreur lors du vote', 'error');
                }
            })
            .catch(error => {
                console.error('Erreur:', error);
                showNotification('Erreur de connexion', 'error');
            });
        });
    });
}

function updateVoteDisplay(targetType, targetId, voteType) {
    // Mettre à jour les boutons de vote visuellement
    const container = document.querySelector(`[data-${targetType}-id="${targetId}"]`);
    if (container) {
        const likeBtn = container.querySelector('.vote-btn[data-type="like"]');
        const dislikeBtn = container.querySelector('.vote-btn[data-type="dislike"]');
        
        // Réinitialiser les styles
        likeBtn.classList.remove('active');
        dislikeBtn.classList.remove('active');
        
        // Activer le bouton cliqué
        if (voteType === 'like') {
            likeBtn.classList.add('active');
        } else {
            dislikeBtn.classList.add('active');
        }
    }
}

// === SYSTÈME DE MODÉRATION ===
function initModeration() {
    // Boutons de bannissement
    document.querySelectorAll('.ban-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            
            const userId = this.dataset.userId;
            const username = this.dataset.username;
            
            if (confirm(`Êtes-vous sûr de vouloir bannir ${username} ?`)) {
                const reason = prompt('Raison du bannissement:');
                if (reason) {
                    banUser(userId, reason);
                }
            }
        });
    });
    
    // Boutons de suppression
    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            
            const itemType = this.dataset.type;
            const itemId = this.dataset.id;
            
            if (confirm(`Êtes-vous sûr de vouloir supprimer ce ${itemType} ?`)) {
                deleteItem(itemType, itemId);
            }
        });
    });
    
    // Boutons de promotion
    document.querySelectorAll('.promote-btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            e.preventDefault();
            
            const userId = this.dataset.userId;
            const username = this.dataset.username;
            const currentRole = this.dataset.currentRole;
            
            showRolePromotionModal(userId, username, currentRole);
        });
    });
}

function banUser(userId, reason) {
    fetch('/api/ban', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            user_id: userId,
            reason: reason
        })
    })
    .then(response => {
        if (response.ok) {
            showNotification('Utilisateur banni avec succès', 'success');
            location.reload();
        } else {
            showNotification('Erreur lors du bannissement', 'error');
        }
    })
    .catch(error => {
        console.error('Erreur:', error);
        showNotification('Erreur de connexion', 'error');
    });
}

function deleteItem(type, id) {
    fetch('/api/delete', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            type: type,
            id: id
        })
    })
    .then(response => {
        if (response.ok) {
            showNotification('Élément supprimé avec succès', 'success');
            // Supprimer l'élément de l'interface
            const element = document.querySelector(`[data-${type}-id="${id}"]`);
            if (element) {
                element.remove();
            }
        } else {
            showNotification('Erreur lors de la suppression', 'error');
        }
    })
    .catch(error => {
        console.error('Erreur:', error);
        showNotification('Erreur de connexion', 'error');
    });
}

function showRolePromotionModal(userId, username, currentRole) {
    const roles = {
        1: 'Utilisateur',
        2: 'Professeur',
        3: 'Modérateur',
        4: 'Administrateur'
    };
    
    let roleOptions = '';
    for (let roleId in roles) {
        if (roleId != currentRole) {
            roleOptions += `<option value="${roleId}">${roles[roleId]}</option>`;
        }
    }
    
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.innerHTML = `
        <div class="modal">
            <h3>Changer le rôle de ${username}</h3>
            <p>Rôle actuel: ${roles[currentRole]}</p>
            <select id="newRole">
                <option value="">Choisir un nouveau rôle</option>
                ${roleOptions}
            </select>
            <div class="modal-buttons">
                <button onclick="promoteUser(${userId}, document.getElementById('newRole').value); this.closest('.modal-overlay').remove()">Confirmer</button>
                <button onclick="this.closest('.modal-overlay').remove()">Annuler</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
}

function promoteUser(userId, newRoleId) {
    if (!newRoleId) {
        showNotification('Veuillez sélectionner un rôle', 'warning');
        return;
    }
    
    fetch('/api/promote', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            user_id: userId,
            role_id: newRoleId
        })
    })
    .then(response => {
        if (response.ok) {
            showNotification('Rôle modifié avec succès', 'success');
            location.reload();
        } else {
            showNotification('Erreur lors de la modification du rôle', 'error');
        }
    })
    .catch(error => {
        console.error('Erreur:', error);
        showNotification('Erreur de connexion', 'error');
    });
}

// === VALIDATION DES FORMULAIRES ===
function initFormValidation() {
    // Validation du formulaire de connexion
    const loginForm = document.querySelector('.auth-form');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            const username = this.querySelector('input[name="username"]').value;
            const password = this.querySelector('input[name="password"]').value;
            
            if (username.length < 3) {
                e.preventDefault();
                showNotification('Le nom d\'utilisateur doit contenir au moins 3 caractères', 'error');
                return;
            }
            
            if (password.length < 6) {
                e.preventDefault();
                showNotification('Le mot de passe doit contenir au moins 6 caractères', 'error');
                return;
            }
        });
    }
    
    // Validation du formulaire de création de post
    const postForm = document.querySelector('#create-post-form');
    if (postForm) {
        postForm.addEventListener('submit', function(e) {
            const title = this.querySelector('input[name="title"]').value;
            const content = this.querySelector('textarea[name="content"]').value;
            const category = this.querySelector('select[name="category_id"]').value;
            
            if (title.length < 5) {
                e.preventDefault();
                showNotification('Le titre doit contenir au moins 5 caractères', 'error');
                return;
            }
            
            if (content.length < 20) {
                e.preventDefault();
                showNotification('Le contenu doit contenir au moins 20 caractères', 'error');
                return;
            }
            
            if (!category) {
                e.preventDefault();
                showNotification('Veuillez sélectionner une catégorie', 'error');
                return;
            }
        });
    }
}

// === SYSTÈME DE NOTIFICATIONS ===
function initNotifications() {
    // Créer le conteneur de notifications s'il n'existe pas
    if (!document.querySelector('.notifications-container')) {
        const container = document.createElement('div');
        container.className = 'notifications-container';
        document.body.appendChild(container);
    }
}

function showNotification(message, type = 'info') {
    const container = document.querySelector('.notifications-container');
    const notification = document.createElement('div');
    
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            <i class="fas fa-${getNotificationIcon(type)}"></i>
            <span>${message}</span>
            <button class="notification-close" onclick="this.parentElement.parentElement.remove()">×</button>
        </div>
    `;
    
    container.appendChild(notification);
    
    // Animation d'entrée
    setTimeout(() => {
        notification.classList.add('show');
    }, 100);
    
    // Auto-suppression après 5 secondes
    setTimeout(() => {
        notification.classList.add('hide');
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 300);
    }, 5000);
}

function getNotificationIcon(type) {
    const icons = {
        success: 'check-circle',
        error: 'exclamation-circle',
        warning: 'exclamation-triangle',
        info: 'info-circle'
    };
    return icons[type] || 'info-circle';
}

// === FONCTIONNALITÉS SUPPLÉMENTAIRES ===

// Marquer une réponse comme solution
function markAsSolution(commentId, postId) {
    if (confirm('Marquer cette réponse comme solution ?')) {
        fetch('/api/solution', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: new URLSearchParams({
                comment_id: commentId,
                post_id: postId
            })
        })
        .then(response => {
            if (response.ok) {
                showNotification('Réponse marquée comme solution !', 'success');
                location.reload();
            } else {
                showNotification('Erreur lors du marquage', 'error');
            }
        })
        .catch(error => {
            console.error('Erreur:', error);
            showNotification('Erreur de connexion', 'error');
        });
    }
}

// Signaler un contenu
function reportContent(type, id, reason) {
    fetch('/api/report', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            type: type,
            id: id,
            reason: reason
        })
    })
    .then(response => {
        if (response.ok) {
            showNotification('Signalement envoyé', 'success');
        } else {
            showNotification('Erreur lors du signalement', 'error');
        }
    })
    .catch(error => {
        console.error('Erreur:', error);
        showNotification('Erreur de connexion', 'error');
    });
}

// Recherche en temps réel
function initSearch() {
    const searchInput = document.querySelector('#search-input');
    if (searchInput) {
        let searchTimeout;
        
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            const query = this.value;
            
            if (query.length >= 3) {
                searchTimeout = setTimeout(() => {
                    performSearch(query);
                }, 500);
            }
        });
    }
}

function performSearch(query) {
    fetch(`/api/search?q=${encodeURIComponent(query)}`)
        .then(response => response.json())
        .then(data => {
            displaySearchResults(data);
        })
        .catch(error => {
            console.error('Erreur de recherche:', error);
        });
}

function displaySearchResults(results) {
    const resultsContainer = document.querySelector('#search-results');
    if (resultsContainer) {
        resultsContainer.innerHTML = '';
        
        if (results.length === 0) {
            resultsContainer.innerHTML = '<p>Aucun résultat trouvé</p>';
            return;
        }
        
        results.forEach(result => {
            const resultElement = document.createElement('div');
            resultElement.className = 'search-result';
            resultElement.innerHTML = `
                <h4><a href="/post/${result.id}">${result.title}</a></h4>
                <p>${result.excerpt}</p>
                <small>dans ${result.category}</small>
            `;
            resultsContainer.appendChild(resultElement);
        });
    }
}

// === STYLES CSS POUR LES NOTIFICATIONS ET MODALES ===
const dynamicStyles = `
.notifications-container {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 1000;
    max-width: 400px;
}

.notification {
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    margin-bottom: 10px;
    transform: translateX(100%);
    transition: transform 0.3s ease;
}

.notification.show {
    transform: translateX(0);
}

.notification.hide {
    transform: translateX(100%);
}

.notification-content {
    padding: 16px;
    display: flex;
    align-items: center;
    gap: 12px;
}

.notification-success {
    border-left: 4px solid var(--success-color);
}

.notification-error {
    border-left: 4px solid var(--danger-color);
}

.notification-warning {
    border-left: 4px solid var(--warning-color);
}

.notification-info {
    border-left: 4px solid var(--info-color);
}

.notification-close {
    background: none;
    border: none;
    font-size: 18px;
    cursor: pointer;
    margin-left: auto;
    color: #999;
}

.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.modal {
    background: white;
    padding: 24px;
    border-radius: 8px;
    max-width: 400px;
    width: 90%;
}

.modal h3 {
    margin-bottom: 16px;
    color: var(--text-primary);
}

.modal select {
    width: 100%;
    padding: 8px;
    margin: 16px 0;
    border: 1px solid var(--border-color);
    border-radius: 4px;
}

.modal-buttons {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 20px;
}

.modal-buttons button {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.modal-buttons button:first-child {
    background: var(--primary-color);
    color: white;
}

.modal-buttons button:last-child {
    background: var(--bg-tertiary);
    color: var(--text-primary);
}

.vote-btn {
    background: none;
    border: 1px solid var(--border-color);
    padding: 6px 12px;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-right: 8px;
}

.vote-btn:hover {
    background: var(--bg-secondary);
}

.vote-btn.active {
    background: var(--primary-color);
    color: white;
    border-color: var(--primary-color);
}
`;

// Ajouter les styles dynamiques
const styleSheet = document.createElement('style');
styleSheet.textContent = dynamicStyles;
document.head.appendChild(styleSheet); 