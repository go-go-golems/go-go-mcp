// Simple AI Inference Test
// This script tests the basic AI inference functionality using Geppetto APIs

function runInferenceTest() {
    console.log("=== AI Inference Test ===");

    try {
        // Test 1: Check if Geppetto APIs are available
        console.log("1. Checking API availability...");
        
        if (typeof Conversation === 'undefined') {
            console.error("❌ Conversation API not available");
            return;
        }
        
        if (typeof ChatStepFactory === 'undefined') {
            console.error("❌ ChatStepFactory API not available");
            return;
        }
    
    console.log("✅ Geppetto APIs are available");
    
    // Test 2: Create conversation and chat step
    console.log("2. Creating conversation and chat step...");
    
    const conversation = new Conversation();
    const factory = new ChatStepFactory();
    const chatStep = factory.newStep();
    
    console.log("✅ Conversation and ChatStep created");
    
    // Test 3: Set up a simple conversation
    console.log("3. Setting up conversation...");
    
    conversation.addMessage("system", "You are a helpful AI assistant. Always respond with exactly one sentence.");
    conversation.addMessage("user", "What is 2 + 2? Please be brief.");
    
    const prompt = conversation.getSinglePrompt();
    console.log("📝 Conversation prompt:", prompt);
    
    // Test 4: Test synchronous inference
    console.log("4. Testing synchronous inference...");
    
    try {
        const syncResult = chatStep.startBlocking(conversation);
        console.log("✅ Sync inference result:", syncResult);
        
        // Add the response to conversation for context
        conversation.addMessage("assistant", syncResult);
        
    } catch (syncError) {
        console.error("❌ Sync inference failed:", syncError.message);
    }
    
    // Test 5: Test asynchronous inference
    console.log("5. Testing async inference...");
    
    // Add another user message
    conversation.addMessage("user", "What color is the sky? One sentence only.");
    
    chatStep.startAsync(conversation)
        .then(asyncResult => {
            console.log("✅ Async inference result:", asyncResult);
            conversation.addMessage("assistant", asyncResult);
            
            // Test 6: Test streaming inference
            console.log("6. Testing streaming inference...");
            
            conversation.addMessage("user", "Tell me a very short joke in one sentence.");
            
            let streamResult = "";
            const cancel = chatStep.startWithCallbacks(conversation, {
                onResult: (chunk) => {
                    streamResult += chunk;
                    console.log("📡 Stream chunk:", chunk);
                },
                onError: (error) => {
                    console.error("❌ Streaming error:", error);
                },
                onDone: () => {
                    console.log("✅ Streaming complete. Full result:", streamResult);
                    conversation.addMessage("assistant", streamResult);
                    
                    // Final test summary
                    console.log("\n=== Test Summary ===");
                    const messages = conversation.getMessages();
                    console.log(`Total messages in conversation: ${messages.length}`);
                    console.log("✅ All inference tests completed successfully!");
                    console.log("=== End Test ===");
                }
            });
            
        })
        .catch(asyncError => {
            console.error("❌ Async inference failed:", asyncError.message);
        });
    
    } catch (error) {
        console.error("❌ Test failed with error:", error.message);
        console.error("Stack trace:", error.stack);
    }
}

// Run the test
runInferenceTest();

// Add API endpoint for testing via HTTP
app.get("/test-inference", (req, res) => {
    res.json({
        message: "AI inference test script loaded",
        timestamp: new Date().toISOString(),
        apis_available: {
            conversation: typeof Conversation !== 'undefined',
            chatStepFactory: typeof ChatStepFactory !== 'undefined'
        }
    });
});

console.log("🌐 Test endpoint available at /test-inference");
