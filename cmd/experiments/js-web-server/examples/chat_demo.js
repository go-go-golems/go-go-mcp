console.log('=== Testing ChatStepFactory ===');

// Test ChatStepFactory availability
if (typeof ChatStepFactory !== 'undefined') {
    console.log('ChatStepFactory available');
    
    try {
        // Create factory instance
        const factory = new ChatStepFactory();
        console.log('ChatStepFactory instance created');
        
        // Create a new step
        const chatStep = factory.newStep();
        console.log('Chat step created:', typeof chatStep);
        
        // List available methods on the step
        const stepMethods = Object.getOwnPropertyNames(chatStep);
        console.log('Chat step methods:', stepMethods);
        
        // Create a conversation for testing
        const conv = new Conversation();
        conv.addMessage("system", "You are a helpful assistant.");
        conv.addMessage("user", "Say hello in one word.");
        
        console.log('Test conversation created with prompt:', conv.getSinglePrompt());
        
        // Note: We can't actually run the step without proper LLM configuration
        // but we can verify the structure is correct
        console.log('Chat step integration test complete');
        
    } catch (error) {
        console.error('ChatStepFactory error:', error.message);
    }
} else {
    console.log('ChatStepFactory not available');
}

console.log('=== Chat Demo Complete ===');
