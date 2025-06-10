console.log('Testing Conversation');
console.log('Conversation type:', typeof Conversation);

if (typeof Conversation !== 'undefined') {
    const conv = new Conversation();
    console.log('Conversation created:', typeof conv);
    console.log('Available methods:', Object.getOwnPropertyNames(conv));
    
    // Try to call addMessage (lowercase due to field name mapper)
    try {
        const msgId = conv.addMessage("user", "Hello, test!");
        console.log('addMessage result:', msgId);
    } catch (error) {
        console.error('addMessage error:', error.message);
    }
} else {
    console.log('Conversation is undefined');
}
