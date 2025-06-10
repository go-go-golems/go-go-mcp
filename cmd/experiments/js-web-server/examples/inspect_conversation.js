console.log('=== Inspecting Conversation ===');
console.log('Conversation type:', typeof Conversation);

if (typeof Conversation !== 'undefined') {
    try {
        const conv = new Conversation();
        console.log('Conversation instance created');
        console.log('Type of conv:', typeof conv);
        console.log('Constructor name:', conv.constructor.name);
        
        // List all properties and methods
        const props = Object.getOwnPropertyNames(conv);
        console.log('Available properties/methods:', props);
        
        // Try to see if there are any methods that contain "message"
        const messageMethods = props.filter(prop => prop.toLowerCase().includes('message'));
        console.log('Methods containing "message":', messageMethods);
        
        // Check for add methods
        const addMethods = props.filter(prop => prop.toLowerCase().includes('add'));
        console.log('Methods containing "add":', addMethods);
        
        // Check prototype
        const protoProp = Object.getOwnPropertyNames(Object.getPrototypeOf(conv));
        console.log('Prototype properties:', protoProp);
        
    } catch (error) {
        console.error('Error creating Conversation:', error.message);
    }
} else {
    console.log('Conversation is undefined');
}

console.log('=== End Inspection ===');
