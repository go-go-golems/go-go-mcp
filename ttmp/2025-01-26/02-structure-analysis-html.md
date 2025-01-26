# HTML Pattern Detection Architecture

## Signature Generation Overview
The signature generation system is designed to identify and match similar HTML structures while being flexible enough to handle variations in content and minor structural differences. It uses a layered approach, starting from strict matching and progressively becoming more lenient based on configuration.

### Core Principles
1. Hierarchical representation of DOM structure
2. Handling of repeating elements
3. Recognition of structural vs. content classes
4. Configurable depth and detail level
5. Balance between specificity and generalization

## Signature Types

### 1. Strict Structural Signature
This signature type focuses purely on HTML tag structure and their immediate relationships. It's useful for finding exact structural matches regardless of content or attributes.

```python
class StrictSignatureGenerator:
    """
    Generates signatures based purely on HTML structure.
    Used for finding exact structural matches with no variation allowed.
    """
    def generate(element: Element, depth: int = 0) -> string:
        # Base structure
        sig = element.tagName
        
        # Handle children
        if element.children and depth < MAX_DEPTH:
            child_sigs = []
            current_group = []
            
            for child in element.children:
                if not current_group or is_similar_structure(child, current_group[0]):
                    current_group.append(child)
                else:
                    # Process previous group
                    child_sigs.append(self._process_group(current_group))
                    current_group = [child]
            
            # Process final group
            if current_group:
                child_sigs.append(self._process_group(current_group))
            
            sig += "(" + "+".join(child_sigs) + ")"
            
        return sig

    def _process_group(self, elements: List[Element]) -> string:
        """
        Processes a group of similar elements into a signature component.
        Handles repeating elements with multiplier notation.
        """
        if len(elements) == 1:
            return self.generate(elements[0])
        else:
            base_sig = self.generate(elements[0])
            return f"{base_sig}*{len(elements)}"
```

### 2. Class-Aware Signature
This signature incorporates CSS classes that appear to be structural rather than decorative. It's more flexible than the strict signature while still maintaining meaningful structural matching.

```python
class ClassAwareSignatureGenerator:
    """
    Generates signatures that include structural CSS classes.
    Distinguishes between structural and decorative classes.
    """
    def __init__(self):
        self.structural_patterns = {
            r'(container|wrapper|section|row|col|grid)': 1.0,
            r'(item|card|panel)': 0.8,
            r'(header|footer|sidebar|content)': 0.9,
            # Weight patterns by likelihood of being structural
        }
    
    def generate(self, element: Element) -> string:
        sig_parts = [element.tagName]
        
        # Filter and sort structural classes
        structural_classes = []
        for class_name in element.classList:
            weight = self._get_structural_weight(class_name)
            if weight > 0.5:  # Threshold for considering structural
                structural_classes.append((class_name, weight))
        
        if structural_classes:
            # Sort by weight and add to signature
            sorted_classes = sorted(structural_classes, key=lambda x: (-x[1], x[0]))
            sig_parts.extend(f".{c[0]}" for c in sorted_classes)
        
        # Process children with structural context
        child_sigs = self._process_children(element)
        if child_sigs:
            sig_parts.append(f"({'+'.join(child_sigs)})")
        
        return "".join(sig_parts)

    def _get_structural_weight(self, class_name: str) -> float:
        """
        Determines how likely a class is to be structural vs decorative.
        Returns weight between 0 and 1.
        """
        # Implementation includes pattern matching and heuristics
```

### 3. Adaptive Signature
This is the most sophisticated signature type, automatically adjusting its matching criteria based on the document structure and pattern frequency.

```python
class AdaptiveSignatureGenerator:
    """
    Generates signatures that adapt to the document's structural patterns.
    Automatically adjusts specificity based on pattern frequency.
    """
    def __init__(self):
        self.pattern_cache = {}
        self.frequency_threshold = 0.1  # Adjustable based on document size
        
    def analyze_document(self, doc: Document):
        """
        Pre-analyzes document to establish pattern baselines.
        Affects subsequent signature generation.
        """
        self.tag_frequencies = self._count_tag_frequencies(doc)
        self.class_patterns = self._analyze_class_patterns(doc)
        self.structure_patterns = self._analyze_common_structures(doc)
        
    def generate(self, element: Element, context: dict = None) -> string:
        if not context:
            context = self._create_context(element)
            
        signature = []
        
        # Add tag with contextual importance
        tag_significance = self._calculate_tag_significance(
            element.tagName, 
            context
        )
        signature.append(self._format_tag(element.tagName, tag_significance))
        
        # Add classes based on structural significance
        structural_classes = self._filter_structural_classes(
            element.classList,
            context
        )
        if structural_classes:
            signature.extend(self._format_classes(structural_classes))
            
        # Process children adaptively
        child_sigs = self._process_children_adaptive(element, context)
        if child_sigs:
            signature.append(self._format_children(child_sigs))
            
        return self._compile_signature(signature)

    def _calculate_tag_significance(self, tag: str, context: dict) -> float:
        """
        Determines how important this tag is for the pattern.
        Uses frequency analysis and structural position.
        """
        # Implementation includes frequency analysis and positional weighting

    def _filter_structural_classes(self, classes: List[str], context: dict) -> List[str]:
        """
        Identifies which classes are significant for structure.
        Uses pattern analysis and context.
        """
        # Implementation includes pattern matching and contextual analysis
```

### Key Supporting Components

```python
class PatternAnalyzer:
    """
    Analyzes pattern frequencies and relationships in the document.
    Used by signature generators to tune their behavior.
    """
    def find_patterns(self, doc: Document) -> Dict[str, Pattern]:
        signatures = {}
        
        # Generate initial signatures
        for element in doc.getElementsByTagName('*'):
            sig = self.generator.generate(element)
            if sig not in signatures:
                signatures[sig] = Pattern(sig)
            signatures[sig].add_occurrence(element)
            
        # Filter and combine related patterns
        return self._consolidate_patterns(signatures)

    def _consolidate_patterns(self, signatures: Dict[str, Pattern]) -> Dict[str, Pattern]:
        """
        Combines related patterns and filters insignificant ones.
        Uses similarity metrics to group related patterns.
        """
        # Implementation includes pattern grouping and filtering
```
