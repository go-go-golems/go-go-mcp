import { useGetModelsQuery } from '../services/api';

const ModelInfo = () => {
  const { data, error, isLoading } = useGetModelsQuery();

  if (isLoading) {
    return (
      <div className="card mb-4">
        <div className="card-header bg-secondary text-white">
          <h5 className="mb-0">Model Information</h5>
        </div>
        <div className="card-body text-center p-4">
          <div className="spinner-border text-primary" role="status">
            <span className="visually-hidden">Loading...</span>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="card mb-4">
        <div className="card-header bg-danger text-white">
          <h5 className="mb-0">Model Information</h5>
        </div>
        <div className="card-body">
          <div className="alert alert-danger mb-0">
            Failed to load model information.
          </div>
        </div>
      </div>
    );
  }

  if (!data) {
    return null;
  }

  return (
    <div className="card mb-4">
      <div className="card-header bg-secondary text-white">
        <h5 className="mb-0">Model Information</h5>
      </div>
      <div className="card-body">
        <h6>Current Model</h6>
        <p className="fw-bold">{data.current_model}</p>
        
        <h6>Description</h6>
        <p>{data.description}</p>
        
        {data.alternatives.length > 0 && (
          <>
            <h6>Alternative Models</h6>
            <ul className="list-group">
              {data.alternatives.map((model) => (
                <li key={model} className="list-group-item">{model}</li>
              ))}
            </ul>
          </>
        )}
      </div>
    </div>
  );
};

export default ModelInfo;