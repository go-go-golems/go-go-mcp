let autoRefreshInterval = null;
let lastGlobalStateValue = '';

async function refreshGlobalState() {
    try {
        const response = await fetch('/admin/globalstate', {
            headers: { 'Accept': 'application/json' }
        });
        const data = await response.text();
        
        document.getElementById('globalStateEditor').value = data;
        lastGlobalStateValue = data;
        
        validateJSON();
        updateLastUpdated();
        
    } catch (error) {
        console.error('Failed to refresh globalState:', error);
        showNotification('Failed to refresh globalState', 'error');
    }
}

async function saveGlobalState() {
    const editor = document.getElementById('globalStateEditor');
    const jsonData = editor.value;
    
    if (!validateJSON()) {
        showNotification('Cannot save invalid JSON', 'error');
        return;
    }
    
    try {
        const formData = new FormData();
        formData.append('globalState', jsonData);
        
        const response = await fetch('/admin/globalstate', {
            method: 'POST',
            body: formData
        });
        
        if (response.ok) {
            lastGlobalStateValue = jsonData;
            showNotification('GlobalState saved successfully', 'success');
            updateLastUpdated();
        } else {
            const error = await response.text();
            showNotification('Failed to save: ' + error, 'error');
        }
    } catch (error) {
        console.error('Failed to save globalState:', error);
        showNotification('Failed to save globalState', 'error');
    }
}

async function resetGlobalState() {
    if (confirm('Are you sure you want to reset globalState to an empty object? This cannot be undone.')) {
        document.getElementById('globalStateEditor').value = '{}';
        await saveGlobalState();
    }
}

function validateJSON() {
    const editor = document.getElementById('globalStateEditor');
    const status = document.getElementById('validationStatus');
    const errorMessage = document.getElementById('errorMessage');
    
    try {
        JSON.parse(editor.value);
        status.textContent = 'Valid JSON';
        status.className = 'validation-status valid';
        errorMessage.textContent = '';
        return true;
    } catch (error) {
        status.textContent = 'Invalid JSON';
        status.className = 'validation-status invalid';
        errorMessage.textContent = ' - ' + error.message;
        return false;
    }
}

function toggleAutoRefresh() {
    const checkbox = document.getElementById('autoRefresh');
    if (checkbox.checked) {
        autoRefreshInterval = setInterval(refreshGlobalState, 5000);
    } else {
        if (autoRefreshInterval) {
            clearInterval(autoRefreshInterval);
            autoRefreshInterval = null;
        }
    }
}

function updateLastUpdated() {
    document.getElementById('lastUpdated').textContent = new Date().toLocaleTimeString();
}

function showNotification(message, type) {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = 'notification ' + type + ' show';
    
    setTimeout(() => {
        notification.classList.remove('show');
    }, 3000);
}

// Validate JSON as user types
document.getElementById('globalStateEditor').addEventListener('input', validateJSON);

// Check for unsaved changes before leaving
window.addEventListener('beforeunload', (e) => {
    const currentValue = document.getElementById('globalStateEditor').value;
    if (currentValue !== lastGlobalStateValue) {
        e.preventDefault();
        e.returnValue = '';
    }
});

// Load initial data
refreshGlobalState();
